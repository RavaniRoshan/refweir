#!/usr/bin/env bash
# Record the containment demo referenced by docs/launch/CONTAINMENT_DEMO.md §6.
#
# This is a launch-asset helper, not part of the product. It records
# examples/containment/run.sh with PACING=1 (which only adds sleeps and timing
# markers), then burns chapter captions, a title card, and an end card.
#
# Requires: vhs, ffmpeg (with libass + libx264), ffprobe, python3.
# Usage:    docs/launch/record-containment.sh [output.mp4]
#
# The default output is docs/launch/assets/containment-demo.mp4.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
OUT="${1:-$ROOT_DIR/docs/launch/assets/containment-demo.mp4}"
OUT_DIR="$(dirname "$OUT")"
mkdir -p "$OUT_DIR"

for cmd in vhs ffmpeg ffprobe python3; do
    command -v "$cmd" >/dev/null 2>&1 || {
        echo "error: '$cmd' is required" >&2
        exit 1
    }
done

WORK="$(mktemp -d "${TMPDIR:-/tmp}/refweir-record.XXXXXX")"
cleanup() {
    if [ "${KEEP_WORK:-0}" = "1" ]; then
        echo "workdir kept at: $WORK" >&2
        return
    fi
    rm -rf "$WORK"
}
trap cleanup EXIT

TAPE_START="$WORK/tape-start.txt"
MARKERS="$WORK/markers.txt"
RAW="$WORK/run-raw.mp4"
CAPTIONED="$WORK/run-captioned.mp4"
TITLE="$WORK/title.mp4"
END="$WORK/end.mp4"
CAPTIONS_ASS="$WORK/captions.ass"
TITLE_ASS="$WORK/title.ass"
END_ASS="$WORK/end.ass"
TAPE="$WORK/containment.tape"

cat >"$TAPE" <<EOF
Output run-raw.mp4
Set Shell "bash"
Set FontSize 18
Set Width 1600
Set Height 1000
Set Padding 24
Set Framerate 30
Set CursorBlink false
Set TypingSpeed 0ms
Set Theme {"background":"#0d1117","foreground":"#c9d1d9","cursor":"#58a6ff","selection":"#264f78","black":"#484f58","red":"#ff7b72","green":"#3fb950","yellow":"#d29922","blue":"#58a6ff","purple":"#bc8cff","cyan":"#39c5cf","white":"#b1bac4","brightBlack":"#6e7681","brightRed":"#ffa198","brightGreen":"#56d364","brightYellow":"#e3b341","brightBlue":"#79c0ff","brightPurple":"#d2a8ff","brightCyan":"#56d4dd","brightWhite":"#f0f6fc"}

Hide
Type "cd $ROOT_DIR && export PACING=1 PACING_MARKERS=$MARKERS && date +%s.%N > $TAPE_START && clear"
Enter
Sleep 400ms
Show
Hide
Type "examples/containment/run.sh; for i in \$(seq 1 8); do sleep 1; done"
Enter
Show
Sleep 100s
EOF

echo "==> recording the paced run (about 75 seconds)"
(cd "$WORK" && vhs "$TAPE" </dev/null)

[ -f "$TAPE_START" ] || { echo "error: recording did not write $TAPE_START" >&2; exit 1; }
[ -s "$MARKERS" ] || { echo "error: recording did not write $MARKERS" >&2; exit 1; }

# The tape sleeps much longer than the run so it never cuts the command short.
# Trim at the last marker plus the final hold, before captions are burned.
RUN_SECONDS="$(python3 - "$MARKERS" "$TAPE_START" <<'PY'
import sys

markers_path, start_path = sys.argv[1:3]
with open(start_path, encoding="utf-8") as fh:
    origin = float(fh.read().strip())
end = None
with open(markers_path, encoding="utf-8") as fh:
    for line in fh:
        parts = line.split()
        if len(parts) == 2 and parts[0] == "end":
            end = float(parts[1])
print(f"{max(1.0, end - origin + 6.5):.2f}" if end else "75")
PY
)"

echo "==> burning chapter captions"
python3 - "$MARKERS" "$TAPE_START" "$CAPTIONS_ASS" <<'PY'
import sys

markers_path, start_path, out_path = sys.argv[1:4]
with open(start_path, encoding="utf-8") as fh:
    origin = float(fh.read().strip())

marks = {}
order = []
with open(markers_path, encoding="utf-8") as fh:
    for line in fh:
        parts = line.split()
        if len(parts) != 2:
            continue
        name, ts = parts
        marks[name] = float(ts) - origin + 0.3
        order.append(name)

captions = [
    ("run_start", "One binary. One temp directory. No external services."),
    ("step1", "A signed lease: this repo, these refs, this expiry."),
    ("step2", "The same credential cannot touch main. The hook stops it, not the UI."),
    ("step3", "Expiry is enforced at authentication."),
    ("step4", "Revocation is one tool call. The next push fails."),
    ("audit", "Every decision is in a durable audit log: actor, lease, ref, reason."),
    ("end", "Containment holds — and the audit trail is durable."),
]

events = []
for name, text in captions:
    if name not in marks:
        continue
    start = max(0.0, marks[name])
    later = [marks[n] for n in order if marks.get(n, -1) > start]
    end = min(later) if later else start + 8.0
    if name == "end":
        end = start + 8.0
    if end - start < 0.4:
        end = start + 0.4
    events.append((start, end, text))


def stamp(seconds):
    h = int(seconds // 3600)
    m = int((seconds % 3600) // 60)
    s = seconds % 60
    return f"{h}:{m:02d}:{s:05.2f}"


header = """[Script Info]
ScriptType: v4.00+
PlayResX: 1600
PlayResY: 1080
WrapStyle: 2
ScaledBorderAndShadow: yes

[V4+ Styles]
Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding
Style: Caption,DejaVu Sans,31,&H00FFFFFF,&H000000FF,&H00000000,&H99000000,-1,0,0,0,100,100,0,0,3,11,0,2,60,60,26,1

[Events]
Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text
"""

with open(out_path, "w", encoding="utf-8") as fh:
    fh.write(header)
    for start, end, text in events:
        fh.write(f"Dialogue: 0,{stamp(start)},{stamp(end)},Caption,,0,0,0,,{text}\n")
PY

# Re-encode the raw capture to a constant 30 fps so the cards concat cleanly.
ffmpeg -hide_banner -loglevel error -y -i "$RAW" -t "$RUN_SECONDS" \
    -vf "scale=1536:960,setsar=1,pad=1600:1080:32:0:color=0x0d1117,ass=$CAPTIONS_ASS,fps=30" \
    -c:v libx264 -preset medium -crf 19 -pix_fmt yuv420p "$CAPTIONED"

cat >"$TITLE_ASS" <<'ASS'
[Script Info]
ScriptType: v4.00+
PlayResX: 1600
PlayResY: 1080
WrapStyle: 2
ScaledBorderAndShadow: yes

[V4+ Styles]
Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding
Style: Card,DejaVu Sans,64,&H00F0F6FC,&H000000FF,&H00000000,&H00000000,-1,0,0,0,100,100,0,0,1,0,0,5,80,80,0,1

[Events]
Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text
Dialogue: 0,0:00:00.00,0:00:05.00,Card,,0,0,0,,{\fs64\b1}An AI agent should not hold the keys to main.\N{\fs30\b0\c&H8B949E&}Refweir — a self-hosted Git forge with scoped agent leases
ASS

cat >"$END_ASS" <<'ASS'
[Script Info]
ScriptType: v4.00+
PlayResX: 1600
PlayResY: 1080
WrapStyle: 2
ScaledBorderAndShadow: yes

[V4+ Styles]
Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding
Style: Card,DejaVu Sans,64,&H00F0F6FC,&H000000FF,&H00000000,&H00000000,-1,0,0,0,100,100,0,0,1,0,0,5,80,80,0,1

[Events]
Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text
Dialogue: 0,0:00:00.00,0:00:07.00,Card,,0,0,0,,{\fs64\b1}Containment you can audit.\N{\fs36\b0}Self-host Refweir.\N{\fs30\c&H58A6FF&}refweir.dev
ASS

echo "==> rendering title and end cards"
ffmpeg -hide_banner -loglevel error -y \
    -f lavfi -i "color=c=0x0d1117:s=1600x1080:r=30:d=5" \
    -vf "ass=$TITLE_ASS,setsar=1" -c:v libx264 -preset medium -crf 19 -pix_fmt yuv420p "$TITLE"
ffmpeg -hide_banner -loglevel error -y \
    -f lavfi -i "color=c=0x0d1117:s=1600x1080:r=30:d=7" \
    -vf "ass=$END_ASS,setsar=1" -c:v libx264 -preset medium -crf 19 -pix_fmt yuv420p "$END"

echo "==> assembling $OUT"
ffmpeg -hide_banner -loglevel error -y \
    -i "$TITLE" -i "$CAPTIONED" -i "$END" \
    -filter_complex "[0:v][1:v][2:v]concat=n=3:v=1:a=0[v]" \
    -map "[v]" -c:v libx264 -preset medium -crf 20 -pix_fmt yuv420p \
    -movflags +faststart "$OUT"

DURATION="$(ffprobe -v error -show_entries format=duration -of default=nw=1:nk=1 "$OUT")"
SIZE="$(du -h "$OUT" | cut -f1)"
echo
echo "wrote $OUT (${DURATION}s, ${SIZE})"
echo "Check: the video must show the title, steps 1-4, the audit table, and the end card."
echo "Nothing in it claims model attribution, prompt hashing, or human approval."
