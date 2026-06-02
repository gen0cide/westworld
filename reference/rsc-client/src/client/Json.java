package client;

import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * Tiny self-contained JSON reader for the DEOB render leg ({@link DumpRender}).
 *
 * <p>The deob client classpath carries no JSON library, and adding org.json would
 * pull a jar into the otherwise-pure deob build. The rscdump/1 fixtures are plain
 * objects of strings / numbers / nested objects / number arrays, so a ~120-line
 * recursive-descent reader suffices. NOT a general JSON parser — it covers exactly
 * the value kinds the fixtures use (object, array, string, number, true/false/null).
 */
public final class Json {
    private final Object value; // Map<String,Object> | List<Object> | String | Double | Boolean | null

    private Json(Object v) {
        this.value = v;
    }

    public static Json parse(String s) {
        P p = new P(s);
        Object v = p.value();
        p.ws();
        return new Json(v);
    }

    @SuppressWarnings("unchecked")
    private Map<String, Object> map() {
        return (Map<String, Object>) value;
    }

    public boolean has(String k) {
        return value instanceof Map && map().containsKey(k);
    }

    public boolean isNull(String k) {
        return has(k) && map().get(k) == null;
    }

    public Json obj(String k) {
        return new Json(map().get(k));
    }

    public int i(String k) {
        return (int) Math.round(((Number) map().get(k)).doubleValue());
    }

    public String str(String k) {
        return (String) map().get(k);
    }

    @SuppressWarnings("unchecked")
    public List<Object> arr(String k) {
        return (List<Object>) map().get(k);
    }

    // ---- recursive-descent parser ----
    private static final class P {
        final String s;
        int i;

        P(String s) {
            this.s = s;
        }

        void ws() {
            while (i < s.length() && Character.isWhitespace(s.charAt(i))) i++;
        }

        Object value() {
            ws();
            char c = s.charAt(i);
            switch (c) {
                case '{':
                    return object();
                case '[':
                    return array();
                case '"':
                    return string();
                case 't':
                    i += 4;
                    return Boolean.TRUE;
                case 'f':
                    i += 5;
                    return Boolean.FALSE;
                case 'n':
                    i += 4;
                    return null;
                default:
                    return number();
            }
        }

        Map<String, Object> object() {
            Map<String, Object> m = new LinkedHashMap<>();
            i++; // {
            ws();
            if (s.charAt(i) == '}') {
                i++;
                return m;
            }
            while (true) {
                ws();
                String key = string();
                ws();
                i++; // :
                Object v = value();
                m.put(key, v);
                ws();
                char c = s.charAt(i++);
                if (c == '}') break;
                // c == ','
            }
            return m;
        }

        List<Object> array() {
            List<Object> a = new ArrayList<>();
            i++; // [
            ws();
            if (s.charAt(i) == ']') {
                i++;
                return a;
            }
            while (true) {
                a.add(value());
                ws();
                char c = s.charAt(i++);
                if (c == ']') break;
                // c == ','
            }
            return a;
        }

        String string() {
            StringBuilder sb = new StringBuilder();
            i++; // opening "
            while (true) {
                char c = s.charAt(i++);
                if (c == '"') break;
                if (c == '\\') {
                    char e = s.charAt(i++);
                    switch (e) {
                        case 'n': sb.append('\n'); break;
                        case 't': sb.append('\t'); break;
                        case 'r': sb.append('\r'); break;
                        case 'b': sb.append('\b'); break;
                        case 'f': sb.append('\f'); break;
                        case '/': sb.append('/'); break;
                        case '\\': sb.append('\\'); break;
                        case '"': sb.append('"'); break;
                        case 'u':
                            sb.append((char) Integer.parseInt(s.substring(i, i + 4), 16));
                            i += 4;
                            break;
                        default: sb.append(e);
                    }
                } else {
                    sb.append(c);
                }
            }
            return sb.toString();
        }

        Double number() {
            int start = i;
            while (i < s.length()) {
                char c = s.charAt(i);
                if (c == '-' || c == '+' || c == '.' || c == 'e' || c == 'E' || (c >= '0' && c <= '9')) i++;
                else break;
            }
            return Double.valueOf(s.substring(start, i));
        }
    }
}
