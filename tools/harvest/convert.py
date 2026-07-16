#!/usr/bin/env python3
"""One-time converter: catholic-resources.org lectionary tables -> liturgy/lectionary.json.

Source pages (Felix Just, SJ; 1998/2002 USA Lectionary) live in pages/ next to
this script. Only scripture *citations* are extracted, never reading texts.
Keys must match what liturgy.Compute produces (see liturgy/lectionary.go).
Run from the repo root: python3 tools/harvest/convert.py
"""
import json
import re
import html
import sys
from pathlib import Path

PAGES = Path(__file__).parent / "pages"
CYCLES = ["A", "B", "C"]
WD = {"Sun": 0, "Mon": 1, "Tues": 2, "Tue": 2, "Wed": 3, "Wednes": 3,
      "Thurs": 4, "Thu": 4, "Fri": 5, "Sat": 6, "Satur": 6}
MONTHS = {"Jan": 1, "Feb": 2, "March": 3, "Mar": 3, "April": 4, "Apr": 4, "May": 5,
          "June": 6, "Jun": 6, "July": 7, "Jul": 7, "Aug": 8, "Sept": 9, "Sep": 9,
          "Oct": 10, "Nov": 11, "Dec": 12}

table = {}
conflicts = []


def rows(name):
    raw = (PAGES / name).read_text(encoding="latin-1")
    for r in re.findall(r"<tr[^>]*>(.*?)</tr>", raw, re.S):
        cells = [re.sub(r"\s+", " ", html.unescape(re.sub(r"<[^>]*>", " ", c))).strip()
                 for c in re.findall(r"<t[dh][^>]*>(.*?)</t[dh]>", r, re.S)]
        if cells:
            yield [c.replace("–", "-").replace("—", "-") for c in cells]


def clean(c):
    c = re.sub(r"\((?:diff|new|note|cited[^)]*|see[^)]*)\)", "", c)
    c = re.sub(r"^(?:opt:|†|\*|cf\.|See\b|see\b)+\s*", "", c)
    c = c.split(" - Vg")[0]  # Vulgate/Greek numbering variants: keep the first form
    c = c.replace("[Vulg.]", "")
    c = re.sub(r"[\[\]]", "", c)  # optional-verse brackets: "[6-7,]"
    # episode titles appended to gospel citations: "- Samaritan Woman"
    c = re.sub(r"\s+-\s+[A-Z][A-Za-z&' ]*(?= or |$)", "", c)
    c = re.sub(r"\s*\*+\s*$", "", c)
    c = re.sub(r"\s+", " ", c).strip(" .;,")
    return c


def is_citation(c):
    # "2 Sam 7:4", "3 John 5-8", "Esth C:12", "Eccl/Qoh 1:2-11",
    # "A: Matt 17:1-9 B: ..." (per-cycle gospel kept verbatim)
    c = re.sub(r"^[ABC]: ", "", c)
    return bool(re.match(r"(?:[123] )?[A-Z][a-z]+(?:/[A-Z][a-z]+)? [A-Z]?:?\d", c)) and "Lectionary" not in c


def put(key, cells, has_alleluia=True):
    """cells = cells after the label. Weekday rows carry [first, psalm,
    alleluia, gospel], Sunday rows [first, psalm, second, alleluia, gospel],
    and the fixed-solemnities table [first, psalm, second, gospel] (no
    alleluia column — pass has_alleluia=False). Non-scriptural alleluias
    ("[no bibl. ref.]") must still hold their slot, so drop only
    dates/empties and assign positionally."""
    slots = []
    for c in cells:
        c = c.strip()
        if not c or re.fullmatch(r"[.x*†\s]+", c):
            continue
        if re.fullmatch(r"[\d,/\[\]\s]+(?:or [\d,/\[\]\s]+)?", c):  # date/years columns
            continue
        if re.fullmatch(r"\d+-\w{3,9}-\d+", c):  # trailing "13-July-26" columns
            continue
        if re.fullmatch(r"[Sf]", c):  # Solemnity/feast rank column
            continue
        if re.fullmatch(r"\[.*\]", c):  # editorial notes: "[ omit in '26 ]"
            continue
        slots.append(clean(c))
    if len(slots) < 3:
        return
    r = {"first": slots[0], "psalm": slots[1], "gospel": slots[-1]}
    if not (is_citation(r["first"]) and is_citation(r["gospel"])):
        return
    if len(slots) >= (5 if has_alleluia else 4) and is_citation(slots[2]):
        r["second"] = slots[2]
    if has_alleluia:  # 5 slots: [f, p, s, a, g]; 4 slots: [f, p, a, g]
        a = slots[3] if len(slots) >= 5 else slots[2] if len(slots) == 4 else ""
        if is_citation(a):
            r["alleluia"] = a
    if key in table:
        if table[key] != r:
            conflicts.append((key, table[key], r))
        return
    table[key] = r


def cycle_suffix(label, default="A"):
    m = re.search(r"- ([ABC])\b", label)
    return m.group(1) if m else default


def lect_no(cell):
    m = re.match(r"\[?(\d+)\]?[*†]*$", cell.strip())
    return int(m.group(1)) if m else None


# ---- Sunday volume: number-driven (lectionary numbers are canonical) ----
def sunday_page(name):
    for cells in rows(name):
        n = label = None
        for i, c in enumerate(cells[:3]):
            n = lect_no(c)
            if n is not None:
                label = cells[i + 1] if i + 1 < len(cells) else ""
                rest = cells[i + 2:]
                break
        if n is None or label is None:
            continue
        key = None
        if 1 <= n <= 12:
            key = f"ADV-{(n - 1) // 3 + 1}-0-{CYCLES[(n - 1) % 3]}"
        elif n == 16:
            key = "FEAST-12-25"
        elif n == 17:
            key = "HOLYFAM-" + cycle_suffix(label)
        elif n == 18:
            key = "FEAST-01-01"
        elif n == 20:
            key = "EPIPH"
        elif n == 21:
            key = "BAPTISM-" + cycle_suffix(label)
        elif 22 <= n <= 36:
            key = f"LENT-{(n - 22) // 3 + 1}-0-{CYCLES[(n - 22) % 3]}"
        elif n == 38:
            # Palm Sunday: one row, but the Passion gospel differs per cycle
            # ("A: Matt 26... B: Mark 14... C: Luke 22...")
            for cy in CYCLES:
                cells = [c for c in rest]
                for i, c in enumerate(cells):
                    if "A: " in c:
                        by_cycle = dict(p.split(": ", 1) for p in re.split(r" (?=[ABC]: )", c))
                        cells[i] = by_cycle.get(cy, c)
                put(f"LENT-6-0-{cy}", cells)
            continue
        elif n == 39:
            key = "LENT-6-4"
        elif n == 40:
            key = "LENT-6-5"
        elif n == 42:
            for cy in CYCLES:
                put(f"EASTER-1-0-{cy}", rest)
            continue
        elif 43 <= n <= 57:
            key = f"EASTER-{(n - 43) // 3 + 2}-0-{CYCLES[(n - 43) % 3]}"
        elif n == 58:
            key = "ASCENSION-" + cycle_suffix(label)
        elif 59 <= n <= 61:
            key = f"EASTER-7-0-{CYCLES[n - 59]}"
        elif n == 63:
            key = "PENTECOST-" + cycle_suffix(label)
        elif 64 <= n <= 162:
            key = f"OT-{(n - 64) // 3 + 2}-0-{CYCLES[(n - 64) % 3]}"
        elif 164 <= n <= 172:
            base = ["TRINITY", "CORPUS", "SACREDHEART"][(n - 164) // 3]
            key = f"{base}-{CYCLES[(n - 164) % 3]}"
        elif n >= 500 and "Vigil" not in label:
            m = re.match(r"(\w+)\.?\s+(?:\[\d+\]/)?(\d+):", label)
            if m and m.group(1) in MONTHS:
                key = f"FEAST-{MONTHS[m.group(1)]:02d}-{int(m.group(2)):02d}"
                put(key, rest, has_alleluia=False)
            continue
        if key:
            put(key, rest)


# ---- Weekday volumes: label-driven ----
def weekday_page(name, prefix=None, cycle=""):
    for cells in rows(name):
        label_i = None
        for i, c in enumerate(cells[:4]):
            if re.match(r"(\d+\w{2} Week|Week \d|Octave of Easter|Ash Wednesday|"
                        r"\w+day after (Ash Wed|Epiphany)|Holy Week|December \d|"
                        r"\[?(Jan|Dec)\. \d)", c):
                label_i = i
                break
        if label_i is None:
            continue
        label, rest = cells[label_i], cells[label_i + 1:]
        key = None
        if m := re.match(r"(\d+)\w{2} Week of (Advent|Lent|Easter) - (\w+)", label):
            week, season, day = int(m.group(1)), m.group(2), WD.get(m.group(3)[:5].rstrip("sd"))
            if day is None:
                day = WD.get(m.group(3)[:3])
            key = {"Advent": "ADV", "Lent": "LENT", "Easter": "EASTER"}[season] + f"-{week}-{day}"
        elif m := re.match(r"Week (\d+) - (\w+)", label):
            key = f"OT-{m.group(1)}-{WD[m.group(2)[:3] if m.group(2)[:5] not in WD else m.group(2)[:5]]}-{cycle}"
        elif m := re.match(r"Octave of Easter - (\w+)", label):
            key = f"EASTER-1-{WD[m.group(1)[:3]]}"
        elif label.startswith("Ash Wednesday"):
            key = "LENT-0-3"
        elif m := re.match(r"(\w+)day after Ash Wed", label):
            key = f"LENT-0-{WD[m.group(1)[:3]]}"
        elif m := re.match(r"Holy Week - (Mon|Tues|Wed)\b", label):
            key = f"LENT-6-{WD[m.group(1)[:3]]}"
        elif m := re.match(r"December (\d+)", label):
            key = f"ADV-12-{int(m.group(1)):02d}"
        elif m := re.match(r"\[?Dec\. (2[6-9]|3[01])\]?", label):
            key = f"XMAS-12-{int(m.group(1)):02d}"
        elif m := re.match(r"\[?Jan\. ([2-7])\]?", label):
            key = f"XMAS-01-{int(m.group(1)):02d}"
        elif m := re.match(r"(\w+?)day after Epiphany", label):
            key = f"EPIPH-{WD[m.group(1)[:5] if m.group(1)[:5] in WD else m.group(1)[:3]]}"
        if key:
            put(key, rest)


for p in ["1998USL-Advent", "1998USL-Christmas", "1998USL-Lent", "1998USL-Easter",
          "1998USL-OrdinaryA", "1998USL-OrdinaryB", "1998USL-OrdinaryC",
          "1998USL-Solemnities"]:
    sunday_page(p + ".html")
weekday_page("2002USL-Weekdays-OT-I.html", cycle="I")
weekday_page("2002USL-Weekdays-OT-II.html", cycle="II")
weekday_page("2002USL-Weekdays-Lent.html")
weekday_page("2002USL-Weekdays-Easter.html")
weekday_page("2002USL-Weekdays-AdventChristmas.html")

# ---- Sanctoral calendar: saint names by fixed date (name layer only) ----
def rank_label(r):
    if "USA" in r or "Calif" in r or ":" in r:
        return None  # US-only entries
    if r == "Mem.":
        return "memorial"
    if r == "." or r.startswith("Univ"):
        return "optional memorial"
    if r.startswith("Feast"):
        return "feast"
    return None  # Solemnities are handled by the feasts map in liturgy.go


saints = {}
for cells in rows("2002USL-Sanctoral.html"):
    if len(cells) < 4:
        continue
    m = re.match(r"\[?(\w+)\.? (\d+)", cells[1])
    if not (m and m.group(1) in MONTHS):
        continue
    rank = rank_label(cells[3])
    if not rank:
        continue
    name = re.sub(r"\s+", " ", re.sub(r"\s*\([^)]*\)", "", cells[2])).strip(" ,")
    key = f"{MONTHS[m.group(1)]:02d}-{int(m.group(2)):02d}"
    saints.setdefault(key, []).append({"name": name, "rank": rank})

sanctoral = Path("liturgy/sanctoral.json")
lines = [f"  {json.dumps(k)}: {json.dumps(saints[k])}" for k in sorted(saints)]
sanctoral.write_text("{\n" + ",\n".join(lines) + "\n}\n")
print(f"wrote {sum(len(v) for v in saints.values())} saints on {len(saints)} dates to {sanctoral}")

for key, kept, dropped in conflicts:
    print(f"CONFLICT {key}: kept {kept} != {dropped}", file=sys.stderr)

out = Path("liturgy/lectionary.json")
lines = [f"  {json.dumps(k)}: {json.dumps(table[k])}" for k in sorted(table)]
out.write_text("{\n" + ",\n".join(lines) + "\n}\n")
print(f"wrote {len(table)} entries to {out}")
