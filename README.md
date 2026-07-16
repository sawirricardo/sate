# sate 🍢

**sate** stands for *saat teduh* — Indonesian for a quiet time with God, most
commonly known as a daily devotional. It also happens to be the same word as
*sate* (satay), the skewered street food. Both readings are acceptable;
only one is nutritious for the soul.

Run it and you get today's Catholic liturgical day and Mass readings,
computed entirely offline:

```
✝ Thursday, 16 July 2026 — Thursday of the 15th week in Ordinary Time (Year A/II)
  Optional memorial: Our Lady of Mount Carmel

First Reading — Isa 26:7-9, 12, 16-19
  The way of the just is right, the path of the just is right to walk in...

Responsorial Psalm — Ps 102:13-14ab+15, 16-18, 19-21
  But thou, O Lord, endurest for ever: and thy memorial to all generations...

Alleluia — Matt 11:28
  Come to me all you that labor and are burdened, and I will refresh you.

Gospel — Matt 11:28-30
  Come to me all you that labor and are burdened... For my yoke is sweet
  and my burden light.
```

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/sawirricardo/sate/main/install.sh | sh
```

Installs the right binary for macOS/Linux (amd64/arm64) into `~/.local/bin`
(override with `SATE_INSTALL_DIR`). Windows users: grab the `.exe` from the
[releases page](https://github.com/sawirricardo/sate/releases). With Go
installed, `go install github.com/sawirricardo/sate@latest` works too.

## Usage

```sh
sate                  # today's liturgy and readings
sate 2026-12-25       # any other date
sate version          # build/version info
```

Without an installed bible, sate prints the day and the reading citations
only. Install a translation to get full texts:

```sh
sate bible ls         # available translations
sate bible add dra    # Douay-Rheims (English, full Catholic canon)
sate bible add ayt    # Alkitab Yang Terbuka (Indonesian)
sate bible use ayt    # pick the active translation
sate bible rm dra     # uninstall
```

Translations are stored in `~/.local/share/sate/` (Linux/macOS) or
`%AppData%\sate` (Windows). Everything else — calendar, lectionary,
saints — is embedded in the binary; after `bible add`, sate never
touches the network.

## How it works

- **Calendar**: Easter is computed (Meeus algorithm), and the whole
  liturgical year — seasons, week numbering, A/B/C and I/II reading
  cycles, movable and fixed feasts — derives from it. Epiphany is kept on
  Sunday, the Ascension on Thursday, and Corpus Christi on Sunday, per
  Indonesian practice.
- **Lectionary**: 721 embedded entries mapping every liturgical day to its
  reading citations, converted once from the 1998/2002 USA Lectionary
  tables at [catholic-resources.org](https://catholic-resources.org/Lectionary/)
  (Felix Just, SJ) by `tools/harvest/convert.py`. Citations only — texts
  come from installed translations.
- **Scripture**: the `scripture` package parses citations like
  `Rev 11:19a; 12:1-6a; 10ab` or `2 Cor 5:20-6:2` into verse ranges. The
  Douay-Rheims keeps Vulgate psalm numbering, so the lectionary's Hebrew
  psalm numbers are mapped when reading from it.

Known limits: Holy Saturday shows no readings (there is no Mass that day),
Protestant-canon translations (ayt) fall back to citation-only on
deuterocanonical readings, and memorials are shown by name while keeping
the weekday readings, as the rubrics prescribe.

## Bible translations

| ID  | Translation | Language | Canon | License |
|-----|-------------|----------|-------|---------|
| dra | Douay-Rheims 1899 | English | Catholic (73 books) | Public domain |
| ayt | [Alkitab Yang Terbuka](https://ayt.co/) | Indonesian | Protestant (66 books) | CopyLeft, non-commercial (YLSA/SABDA) |
| web | World English Bible | English | Protestant | Public domain |
| kjv | King James Version 1769 | English | Protestant | Public domain |
| asv | American Standard Version 1901 | English | Protestant | Public domain |
| ylt | Young's Literal Translation 1898 | English | Protestant | Public domain |

Only translations whose licenses permit storing the full text locally are
offered. LAI's Terjemahan Baru is copyrighted and therefore absent.

## Build

```sh
make build      # native binary
make test       # go test ./...
make release    # dist/ binaries for darwin/linux/windows, amd64 + arm64
```

Every push to `main` is released automatically: CI runs the tests, tags the
next patch version, and publishes the binaries. To bump minor/major, tag
before pushing (`make ship V=v0.2.0`).
