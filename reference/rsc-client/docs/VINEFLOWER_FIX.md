# Vineflower `parsing failure!` on obfuscated RSC client — repro, root cause, fix

Target: `rscplus/assets/rsclassic-1091943135.jar` (RuneScape Classic, heavily obfuscated).
Symptom: Vineflower 1.10.1 emits `// $VF: Couldn't be decompiled` with
`java.lang.RuntimeException: parsing failure!` at `DomHelper.parseGraph`.
Baseline failures: **238 methods across 41 of 71 classes** (72 in `client` alone).

## TL;DR
- Root cause = the obfuscator's **bogus, mutually-overlapping `RuntimeException`
  exception-table ranges** whose handler is a dead lone `athrow`. They make
  Vineflower's dominator-based structurer unable to reduce the CFG.
- **No upstream Vineflower version fixes it** (tested 1.10.1, 1.12.0, and the
  `feature/better-multi-entry-exception-deobf` branch — all still 72 in `client`).
- **No CLI flag fixes it** (`old-try-dedup`, `try-loop-fix`, `ignore-invalid-bytecode`,
  `remove-empty-try-catch` all tested).
- **Fix = strip the bogus exception-table entries before decompiling.** Two
  independent implementations, both reduce 238 → **4** failures (41 → 4 files):
  1. ASM jar→jar pre-strip (primary).
  2. Krakatau disassemble → normalize `.catch` → reassemble (cross-check).
- The 4 residual failures are genuine loop/nested-try structuring limits, NOT the
  obfuscation (CFR also fails 1 method, `client.l(int)`, for the same class of reason).

## Environment gotchas
- System default JDK is **26**; Gradle 9.2.1 (wrapper) cannot run on JDK 26, and
  the JDK-21 install here is **JRE-only (no javac)**. Build Vineflower with JDK 17:
  ```
  cd /home/free/code/rsc-hacking/vineflower
  JAVA_HOME=/usr/lib/jvm/java-17-openjdk ./gradlew shadowJar --no-daemon
  # output: build/libs/vineflower-<ver>+local-all.jar   (root project IS the decompiler;
  #         there is NO :vineflower: subproject — use the root `shadowJar` task)
  ```

## Reproduce the failure
```
unzip -o rscplus/assets/rsclassic-1091943135.jar client.class -d /tmp/cls
java -jar deob/tools/vineflower.jar /tmp/cls/client.class /tmp/out/
grep -c "Couldn't be decompiled" /tmp/out/client.java        # -> 72
javap -c -p -classpath /tmp/cls client | less                # see I(int) exception table
```
Clean repro method: `client.private final void I(int)` (first failure). Its
exception table has **48 RuntimeException ranges**: 1 real (`from 5 to 740 using 743`)
+ 47 bogus overlapping ranges whose targets are lone `athrow` insns (offsets 37,48,66…).

## Root cause (mechanism)
Every method is `try { BODY } catch (RuntimeException e) { throw i.a(e,"sig"); }`.
On top of the one real range the obfuscator emits ~N extra ranges that:
- partially **overlap each other with different handlers** (e.g. `15–34→37` and
  `23–45→48` share `[23,34]` but neither nests the other), and
- target dead `athrow` instructions embedded in the body (unreachable; preceded by `goto`).

Vineflower (`MethodProcessor.codeToJava` → `ExceptionDeobfuscator
.handleMultipleEntryExceptionRanges` + `insertDummyExceptionHandlerBlocks` →
`DomHelper.parseGraph`) only handles **multiple-entry** ranges; it does not resolve
**non-nested partially-overlapping** ranges with distinct handlers. After preprocessing,
the handler blocks still cross-protect each other (dumped CFG: handler `B174` protects
handler `B175` and vice-versa), so no block postdominates a try-region → `processStatement`
returns false → `throw new RuntimeException("parsing failure!")`. The logged
`Statement cannot be decomposed although reducible!` confirms it's the
reducible-but-undecomposable path, not irreducible-loop handling.

(To re-dump the failing CFG: a temporary instrument was added at `DomHelper.parseGraph`
guarded by env `VF_DEBUG_DUMP` + `mt.getName()=="I"`, then reverted. See git history
of this checkout if needed.)

## THE FIX

### A) ASM pre-strip (primary)  — tools in /tmp/strip, packaged jars in deob/tools
Removes any try-catch whose handler's first real opcode is `ATHROW`
(the obfuscator's dead rethrow handlers); keeps real handlers (first op = aload/astore/...).
Only the exception table changes, so frames/maxs stay valid (verified, 0 VerifyErrors).

```
# self-contained strip tool (bundles ASM 9.8):
java -jar deob/tools/strip-obf-exceptions.jar <in.jar|.class> <out.jar|out.class>
# e.g.
java -jar deob/tools/strip-obf-exceptions.jar \
     rscplus/assets/rsclassic-1091943135.jar /tmp/rsclassic-stripped.jar
#   -> [strip] classes touched=55 methods touched=546 ranges removed=5371
java -jar deob/tools/vineflower.jar /tmp/rsclassic-stripped.jar /tmp/out_stripped/
```

Source: `/tmp/strip/src/StripObfExceptions.java` (compile with asm-9.8 + asm-tree-9.8
from `~/.gradle/caches/modules-2/.../org.ow2.asm/`).

### A') Drop-in wrapper:  deob/tools/vineflower-fixed.jar
One-step strip+decompile (bundles StripObfExceptions + ASM 9.8 + Vineflower 1.12.0).
Same CLI as Vineflower; it strips the input jar/class to a temp dir first.
```
java -jar deob/tools/vineflower-fixed.jar \
     rscplus/assets/rsclassic-1091943135.jar /tmp/out/
#   -> 4 failures across 4 files (was 238 / 41)
```

### B) Krakatau pipeline (independent cross-check)
Built `krak2` (Rust) at `/tmp/Krakatau/target/release/krak2` (`cargo build --release`).
```
krak2 dis /tmp/cls/client.class -o /tmp/krak_dis              # -> client.j
python3 /tmp/strip/normalize_catch.py /tmp/krak_dis/client.j /tmp/krak_dis/client_norm.j
#   -> removed 2427 bogus .catch directives
krak2 asm /tmp/krak_dis/client_norm.j -o /tmp/krak_client.jar
java -jar deob/tools/vineflower.jar <extracted client.class> /tmp/out_krak/
```
Produces **byte-identical** decompiled `I(int)` to the ASM path. Useful when deeper
bytecode surgery is wanted; ASM path is simpler for bulk jar→jar.

## Correctness check (vs CFR oracle), client.I(int)
Stripped-Vineflower output == CFR semantics:
- identical method-call multiset (39 calls), identical field-access multiset,
- identical catch: `catch (RuntimeException e) { throw i.a(e, il[229] + var1 + ')') }`.

## Impact (full jar, Vineflower 1.10.1)
| metric                | original | after strip |
|-----------------------|----------|-------------|
| failed methods        | 238      | **4**       |
| affected files        | 41 / 71  | **4 / 71**  |
| failures in client.java | 72     | **1**       |

Residual 4 (loop/nested-try structuring, not the obfuscation):
`client.a(int,int,boolean)`, `ba.a(...)`, `ib.a(...)`, `k.b(...)`.

## Recommendation
Insert the ASM strip as a pre-decompile step (use `strip-obf-exceptions.jar`, or the
all-in-one `vineflower-fixed.jar`). Upgrading Vineflower alone does nothing for this jar.
