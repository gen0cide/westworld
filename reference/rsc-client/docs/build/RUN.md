# RUN.md — Build & run the decompiled RSC client

A reproducible, copy-pasteable recipe to **compile the deobfuscated client, stand up a
local OpenRSC server + content host, and launch the client into a live in-game 3D
session** on a headless box.

This is the rev ~233–235 J++ `rsclassic` client (the one decompiled in `../src/`). It
boots, autologins against a local OpenRSC server, holds a stable session, and renders the
live 3D world. See [FUNCTIONING_CLIENT.md](FUNCTIONING_CLIENT.md) for the achievement
summary and honest status.

> All paths are absolute. Commands assume a Linux host. The proven environment used
> `Xvfb :99` for a headless display.

---

## 0. Prerequisites

| Requirement | Why | Note |
|---|---|---|
| **JDK 17** (`/usr/lib/jvm/java-17-openjdk`) | The client uses `java.applet.Applet` (`GameShell extends Applet`). | **Do NOT use JDK 21+/26** — `java.applet` was removed; the tree will not compile/run there. JDK 17 still ships it (with a deprecation warning, which is expected and harmless). |
| **Xvfb** (`/usr/bin/Xvfb`) | Headless X display for the AWT window. | A real `$DISPLAY` works too. |
| **Python 3** | Static content host (`content-host.py`). | stdlib only. |
| **Ant 1.10.5** (vendored) | Builds + launches the OpenRSC server. | At `/home/free/code/rsc-hacking/openrsc/Portable_Windows/apache-ant-1.10.5`; invoked by `start-server.sh`. |
| **ImageMagick `import`** (optional) | Capture an Xvfb screenshot. | The client also has a built-in framebuffer-dump hook (no AWT) — see §7. |

Set JDK 17 for the shell:

```sh
export JAVA_HOME=/usr/lib/jvm/java-17-openjdk
export PATH="$JAVA_HOME/bin:$PATH"
java -version    # expect openjdk 17.x
```

---

## 1. Compile the client (M0)

The tree compiles **0 errors / 71 classes** under JDK 17. The two `Applet`-removal
warnings are expected.

```sh
export JAVA_HOME=/usr/lib/jvm/java-17-openjdk
export PATH="$JAVA_HOME/bin:$PATH"
cd /home/free/code/rsc-hacking/westworld

# collect sources -> compile into a FRESH build dir
find reference/rsc-client/src/client -name '*.java' > /tmp/srcs.txt
rm -rf /tmp/deob-run
javac -d /tmp/deob-run -encoding UTF-8 @/tmp/srcs.txt

# sanity: 71 classes, 0 errors
find /tmp/deob-run -name '*.class' | wc -l    # -> 71
```

> Use a **fresh build dir per launch** (`rm -rf` first). Stale `.class` files from a prior
> build are the most common cause of confusing runtime behaviour.

---

## 2. Stand up the OpenRSC server

The server is a separate checkout at `/home/free/code/rsc-hacking/openrsc/server`, built
with JDK 17 via vendored Ant. Build artifacts (`core.jar`, `plugins.jar`), the config
(`westworld.conf`), the RSA keypair (`server.pem`/`client.pem`), and the SQLite DB
(`inc/sqlite/westworld.db`) all persist on disk.

**Start it (background):**

```sh
/tmp/rsc-run/start-server.sh   # see "launcher contents" below
```

`start-server.sh` is:

```sh
#!/bin/sh
export JAVA_HOME=/usr/lib/jvm/java-17-openjdk
export ANT_HOME=/home/free/code/rsc-hacking/openrsc/Portable_Windows/apache-ant-1.10.5
export PATH="$JAVA_HOME/bin:$ANT_HOME/bin:$PATH"
cd /home/free/code/rsc-hacking/openrsc/server || exit 9
exec "$JAVA_HOME/bin/java" -cp "$ANT_HOME/lib/ant-launcher.jar" -Dant.home="$ANT_HOME" \
  org.apache.tools.ant.launch.Launcher runserver -DconfFile=westworld -DcoloredLogging=false
```

**Wait until it is listening** before launching the client:

```sh
# server is up when this prints LISTEN on 43594
ss -ltn | grep 43594
# or grep the run log for: "Game world is now online on TCP port 43594!"
```

**Endpoints / config:**

- **Game TCP socket:** `127.0.0.1:43594`  ← the client connects here.
- DB: **SQLite** at `/home/free/code/rsc-hacking/openrsc/server/inc/sqlite/westworld.db`
  (no MySQL). Config: `westworld.conf` (`enforce_custom_client_version: false`,
  version gate accepts rev-235). Logins from `127.0.0.1` bypass rate/IP limits.

### The `deobtest` account

The autologin uses **`deobtest` / `deobpass`**. The row lives in the SQLite DB above
(password stored bcrypt-hashed). If you start from a fresh DB with no player row, either
register over the wire (first connect takes the in-protocol register-then-login path) or
insert a `players` row directly. The current DB already has the `deobtest` account
(bcrypt hash recorded at `/tmp/rsc-run/deobtest-hash.txt`).

### RSA key (must match)

The client encrypts the login block with the server's **512-bit RSA public key**. The
modulus is baked into the launcher (`Boot.java`, see §5) and must match the live server's
`client.pem`:

- exponent: `65537`
- modulus (decimal):
  `8470727801174954902989859055344934434282083179399207801708507751976321325965228952554034824402302678046886295251980280826867546707365065713308009848924031`

> **WARNING:** deleting the `.pem` files makes the server regenerate a *new* key on next
> boot (the modulus changes) — you would then have to update `Boot.SERVER_RSA_MODULUS` to
> match. Keep the existing `.pem` files.

---

## 3. Stand up the content host

The client fetches its 12 content packs + CRC manifest over plain HTTP on first boot
(`CacheUpdater.downloadAndVerifyCrcs`, which also fills the `_junkArray` CRC table that
`loadResource` needs). A vanilla `python3 -m http.server` will **not** work — it 404s the
cache-buster `contentcrcs<hex-millis>` path. Use the bundled handler.

**Port:** the client computes the content-host port as **nodeId + 7000**, and we launch
with `nodeId=0`, so the host must run on **port 7000** at `127.0.0.1`.

```sh
python3 /tmp/rsc-run/content-host.py 7000   # background; serves /tmp/rsc-run/cache verbatim
```

`content-host.py` serves `/tmp/rsc-run/cache/` verbatim and collapses any `/contentcrcs*`
request to the on-disk `contentcrcs` manifest. The staged cache contains all 12 packs
(`content0..11`), `contentcrcs`, and `jagex.jag` (= content index 3, the logo archive) —
all CRCs consistent with the rev-235 set. All 12 indices must resolve to reach the world;
they are all present.

> Port collision note: the game socket is also 43594 in some legacy scripts, but the
> **content host runs on 7000** for this launch recipe (nodeId 0 + 7000). Do not put the
> content host on 43594 — that is the game socket.

---

## 4. Headless display

```sh
Xvfb :99 -screen 0 1024x768x24 &
export DISPLAY=:99
```

(Skip if you have a real display.)

---

## 5. The launcher: `client.Boot`

`Boot` is a thin standalone entrypoint
(`reference/rsc-client/src/client/Boot.java`) that injects the runtime ADDs, then calls
`Mudclient.main`:

1. **RSA override** — sets `BitBuffer.RSA_MODULUS` and `FontBuilder.rsaPublicExponent` to
   the live server key (both are `public static`), *before* any login.
2. **Host/port** — the host/port is patched in-source in `Mudclient.setLoaderApplet`'s
   `LOCAL_CIPHER` branch to point `serverHost = "127.0.0.1"` and the login port to
   **43594**.
3. **Content** — `Mudclient.main` computes the content-host port as `nodeId + 7000`, so
   passing `nodeId "0"` makes the client fetch packs from `http://127.0.0.1:7000/` with
   `doUpdate > 20` (which runs the CacheUpdater and fills `_junkArray`).

```sh
# Boot args: <nodeId> <mode>   ->   "0" "live"
java -cp /tmp/deob-run client.Boot 0 live
```

### Env-gated bring-up hooks (headless automation)

These let the headless session drive itself with no keyboard/mouse. All are
**env-var-gated** and one-shot/guarded — they send only what the real client would. Set
them in the launch environment:

| Env var | Effect |
|---|---|
| `RSC_AUTOLOGIN_USER` / `RSC_AUTOLOGIN_PASS` | One-shot autologin via the real `loginUser()` handshake (set to `deobtest` / `deobpass`). |
| `RSC_AUTO_APPEARANCE` | Auto-submits the default appearance so the character-design screen clears without a click. |
| `RSC_AUTO_WALK` | Issues real `WALK_TO_POINT` packets on a rectangular tour around spawn (exercises `World.route`, region streaming, camera-follow, op-191 coord updates). Optional. |
| `RSC_AUTO_TABS` | Cycles the active main UI tab so every wired tab renderer runs each session. Optional. |
| `RSC_FBUFFER_DUMP` | After N in-game frames (`RSC_FBUFFER_FRAMES`, default 100), dumps `Surface.pixels` straight to a PNG via `BufferedImage`+`ImageIO` — **no AWT/Xvfb flush** — showing exactly what the renderer produced. Writes `fbuffer.png` (Welcome box up) then `live3d.png` (box dismissed) into `RSC_FBUFFER_DIR` (default `/tmp/rsc-run`). |

---

## 6. Launch recipe (the proven one): `systemd-run`

**Why not just run `java` in the foreground/background from the shell:** a foregrounded /
shell-job JVM here gets **SIGURG-killed** (the harness/job-control delivers SIGURG, which
takes the JVM down). Launching the JVM under a transient **`systemd --user`** unit detaches
it cleanly and survives. Use a **fresh build dir + a unique unit name per launch**.

```sh
# (services from §2/§3/§4 already running: server on 43594, content host on 7000, Xvfb :99)
mkdir -p /tmp/rsc-run

systemd-run --user --unit=rsc-deob-run1 \
  -p StandardOutput=file:/tmp/rsc-run/run1.log \
  -p StandardError=file:/tmp/rsc-run/run1.err \
  --setenv=DISPLAY=:99 \
  --setenv=JAVA_HOME=/usr/lib/jvm/java-17-openjdk \
  --setenv=RSC_AUTOLOGIN_USER=deobtest \
  --setenv=RSC_AUTOLOGIN_PASS=deobpass \
  --setenv=RSC_AUTO_APPEARANCE=1 \
  --setenv=RSC_FBUFFER_DUMP=1 \
  /usr/lib/jvm/java-17-openjdk/bin/java -cp /tmp/deob-run client.Boot 0 live
```

Add `--setenv=RSC_AUTO_WALK=1 --setenv=RSC_AUTO_TABS=1` to drive movement + tab cycling.

**Watch / stop:**

```sh
# follow the client's stdout
tail -f /tmp/rsc-run/run1.log
# you should see: [Boot] RSA modulus bitlen=512 ... then login ... then
# [RSC_FBUFFER_DUMP] wrote /tmp/rsc-run/fbuffer.png ... non-black viewport pixels=...

# stop the run (always stop before relaunching; use a NEW unit name for the next launch)
systemctl --user stop rsc-deob-run1
```

> Each launch: **fresh `/tmp/deob-run` build + a unique `--unit` name** (`rsc-deob-run2`,
> …). Reusing a unit name or a stale build dir is the usual source of confusion.

---

## 7. Capture a frame

Two ways. Prefer the framebuffer dump — it shows exactly what the renderer produced,
independent of any AWT/Xvfb window-flush artifact.

**A. Built-in framebuffer dump (no AWT):** set `RSC_FBUFFER_DUMP=1` (as above). After
~100 in-game frames it writes:
- `/tmp/rsc-run/fbuffer.png` — composited frame as-is (server "Welcome" box still up)
- `/tmp/rsc-run/live3d.png`  — Welcome box dismissed, clean 3D viewport

**B. Xvfb root screenshot:**

```sh
DISPLAY=:99 import -window root /tmp/rsc-run/frame.png
```

---

## 8. One-shot end-to-end (copy-paste)

```sh
export JAVA_HOME=/usr/lib/jvm/java-17-openjdk
export PATH="$JAVA_HOME/bin:$PATH"
export DISPLAY=:99
cd /home/free/code/rsc-hacking/westworld

# 1. compile (fresh)
find reference/rsc-client/src/client -name '*.java' > /tmp/srcs.txt
rm -rf /tmp/deob-run && javac -d /tmp/deob-run -encoding UTF-8 @/tmp/srcs.txt
find /tmp/deob-run -name '*.class' | wc -l        # expect 71

# 2. display + services
pgrep -f "Xvfb :99" >/dev/null || (Xvfb :99 -screen 0 1024x768x24 &)
/tmp/rsc-run/start-server.sh &                    # wait for LISTEN on 43594
python3 /tmp/rsc-run/content-host.py 7000 &       # content host on 7000
until ss -ltn | grep -q 43594; do :; done         # block until server is up

# 3. launch the client under a transient systemd unit
systemd-run --user --unit=rsc-deob-run1 \
  -p StandardOutput=file:/tmp/rsc-run/run1.log \
  -p StandardError=file:/tmp/rsc-run/run1.err \
  --setenv=DISPLAY=:99 \
  --setenv=JAVA_HOME=/usr/lib/jvm/java-17-openjdk \
  --setenv=RSC_AUTOLOGIN_USER=deobtest --setenv=RSC_AUTOLOGIN_PASS=deobpass \
  --setenv=RSC_AUTO_APPEARANCE=1 --setenv=RSC_FBUFFER_DUMP=1 \
  /usr/lib/jvm/java-17-openjdk/bin/java -cp /tmp/deob-run client.Boot 0 live

# 4. watch, then capture
tail -f /tmp/rsc-run/run1.log                     # wait for [RSC_FBUFFER_DUMP] wrote ...
# -> /tmp/rsc-run/fbuffer.png  and  /tmp/rsc-run/live3d.png

# 5. stop
systemctl --user stop rsc-deob-run1
```

---

## Troubleshooting

| Symptom | Cause / fix |
|---|---|
| `package java.applet does not exist` / won't compile | Wrong JDK. Use **17**, not 21+/26. |
| JVM dies instantly, empty log | Foregrounded JVM got SIGURG-killed. Launch via `systemd-run` (§6). |
| Stuck on title / asset 404s | Content host not up on **7000**, or using vanilla `http.server` (404s `/contentcrcs*`). Use `content-host.py`. |
| Login rejected / connection refused | Server not listening on 43594, or RSA modulus mismatch (regenerated `.pem`). Confirm `ss -ltn \| grep 43594` and that `Boot.SERVER_RSA_MODULUS` matches `client.pem`. |
| Weird behaviour after a code change | Stale `.class` files. `rm -rf /tmp/deob-run` and recompile; use a unique `--unit` name. |
| `error in sprite clipping routine` spam | Authentic vanilla RSC diagnostic (swallowed `catch`), data-dependent and transient — not a crash. See FUNCTIONING_CLIENT.md. |
