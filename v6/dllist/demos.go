package dllist

import (
	"NintendoChannel/common"
	"NintendoChannel/constants"
	"NintendoChannel/v6/info"
	"bytes"
	"fmt"
	"hash/crc32"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf16"

	"github.com/kaitai-io/kaitai_struct_go_runtime/kaitai"
)

type DemoTable struct {
	ID            uint32
	Title         [31]uint16
	Subtitle      [31]uint16
	TitleID       uint32
	CompanyOffset uint32
	RemovalYear   uint16
	RemovalMonth  uint8
	RemovalDay    uint8
	_             uint32
	RatingID      uint8
	IsNew         uint8
	_             uint8
	_             [205]byte
}

type DemoEntry struct {
	ID       uint32
	Title    string
	Subtitle string
}

var RegionalDemos = map[constants.Region][]DemoEntry{
	constants.NTSC: {
		{Title: "100 Classic Books"},
		{Title: "America's Test Kitchen", Subtitle: "Pots de Crème Demo"},
		{Title: "America's Test Kitchen", Subtitle: "Roasted Red Potatoes Demo"},
		{Title: "Art Academy"},
		{Title: "Big Brain Academy"},
		{Title: "Brain Age Demo"},
		{Title: "Brain Age 2 Demo"},
		{Title: "Clubhouse Games", Subtitle: "(Darts)"},
		{Title: "Clubhouse Games", Subtitle: "(Turncoat)"},
		{Title: "Crosswords DS-Anagrams Demo"},
		{Title: "Crosswords DS-Crosswords Demo"},
		{Title: "Crosswords DS-Wordsearch Demo"},
		{Title: "Custom Robo Arena"},
		{Title: "Elite Beat Agents"},
		{Title: "Flash Focus Demo"},
		{Title: "Flipnote Studio Theater"},
		{Title: "Fossil Fighters", Subtitle: "Cleaning Demo"},
		{Title: "Fossil Fighters", Subtitle: "Gift Fossil (Air)"},
		{Title: "Fossil Fighters", Subtitle: "Gift Fossil (Earth)"},
		{Title: "Fossil Fighters", Subtitle: "Gift Fossil (Fire)"},
		{Title: "Fossil Fighters", Subtitle: "Gift Fossil (Neutral)"},
		{Title: "Fossil Fighters", Subtitle: "Gift Fossil (Water)"},
		{Title: "Fossil Fighters: Champions"},
		{Title: "Kirby Mass Attack"},
		{Title: "Magnetica"},
		{Title: "Mario Kart DS"},
		{Title: "Mario vs. Donkey Kong 2:", Subtitle: "March of the Minis"},
		{Title: "Mario vs. Donkey Kong:", Subtitle: "Mini-Land Mayhem!"},
		{Title: "Meteos"},
		{Title: "Metroid Prime Hunters"},
		{Title: "Mystery Case Files:", Subtitle: "MillionHeir"},
		{Title: "Personal Trainer: Cooking", Subtitle: "Lasagna Demo"},
		{Title: "Personal Trainer: Cooking", Subtitle: "Mac & Cheese Demo"},
		{Title: "Personal Trainer: Math"},
		{Title: "Picross 3D"},
		{Title: "Picross DS"},
		{Title: "Planet Puzzle League"},
		{Title: "Pokémon Mystery Dungeon:", Subtitle: "Explorers of Sky Demo"},
		{Title: "Pokémon Ranger: Guardian Signs", Subtitle: "Demo"},
		{Title: "Pokémon Trozei!"},
		{Title: "Rhythm Heaven"},
		{Title: "Tetris DS", Subtitle: "(Push)"},
		{Title: "Tetris DS", Subtitle: "(Standard)"},
		{Title: "True Swing Golf"},
	},
	constants.PAL:   {},
	constants.Japan: {},
}

func PopulateDemos() {
	dl := NewNinchDllist()
	err := dl.Read(kaitai.NewStream(bytes.NewReader(constants.DLList)), nil, dl)
	common.CheckError(err)

	origDemos, err := dl.DemosTable()
	common.CheckError(err)

	origIDByKey := map[string]uint32{}
	for _, d := range origDemos {
		key := normalizeTitle(strings.TrimSpace(d.Title + " " + d.Subtitle))
		if key != "" {
			origIDByKey[key] = d.Id
		}
	}

	crcTable := crc32.MakeTable(crc32.IEEE)
	for region, entries := range RegionalDemos {
		for idx := range entries {
			if entries[idx].ID != 0 {
				continue
			}
			full := strings.TrimSpace(entries[idx].Title + " " + entries[idx].Subtitle)
			if id, ok := origIDByKey[normalizeTitle(full)]; ok {
				entries[idx].ID = id
			} else {
				entries[idx].ID = crc32.Checksum([]byte(fmt.Sprintf("%d:%s", region, full)), crcTable)
			}
		}
		RegionalDemos[region] = entries
	}
}

func decodeUTF16Z(u []uint16) string {
	n := 0
	for n < len(u) && u[n] != 0 {
		n++
	}
	return string(utf16.Decode(u[:n]))
}

func (l *List) findTitleForDemo(demo DemoEntry) *TitleTable {
	combined := strings.TrimSpace(demo.Title + " " + demo.Subtitle)
	stripped := demoSuffix.ReplaceAllString(combined, "")
	baseStripped := demoSuffix.ReplaceAllString(strings.TrimSpace(demo.Title), "")

	candidates := []string{
		normalizeTitle(combined),
		normalizeTitle(stripped),
		normalizeTitle(demo.Title),
		normalizeTitle(baseStripped),
	}

	var fallback *TitleTable
	for idx := range l.TitleTable {
		t := &l.TitleTable[idx]
		tName := strings.TrimSpace(decodeUTF16Z(t.TitleName[:]) + " " + decodeUTF16Z(t.Subtitle[:]))
		tBase := decodeUTF16Z(t.TitleName[:])
		nFull := normalizeTitle(tName)
		nBase := normalizeTitle(tBase)

		matched := false
		for _, want := range candidates {
			if want != "" && (nFull == want || nBase == want) {
				matched = true
				break
			}
		}
		if matched {
			if t.TitleType == constants.NintendoDS || t.TitleType == constants.NintendoDSi || t.TitleType == constants.NintendoDSiWare {
				return t
			}
			if fallback == nil {
				fallback = t
			}
		}
	}
	return fallback
}

func (l *List) MakeDemoTable() {
	l.Header.DemoTableOffset = l.GetCurrentSize()

	for _, demo := range RegionalDemos[l.region] {
		var title [31]uint16
		tempTitle := utf16.Encode([]rune(demo.Title))
		copy(title[:], tempTitle)

		var subtitle [31]uint16
		tempSubtitle := utf16.Encode([]rune(demo.Subtitle))
		copy(subtitle[:], tempSubtitle)

		var titleID uint32
		var ratingID uint8 = 8
		companyOffset := l.Header.CompanyTableOffset

		if matched := l.findTitleForDemo(demo); matched != nil {
			titleID = matched.ID
			if matched.RatingID != 0 {
				ratingID = matched.RatingID
			}
			if matched.CompanyOffset != 0 {
				companyOffset = matched.CompanyOffset
			}
		}
		if ratingID < 8 {
			ratingID = 8
		}

		l.DemoTable = append(l.DemoTable, DemoTable{
			ID:            demo.ID,
			Title:         title,
			Subtitle:      subtitle,
			TitleID:       titleID,
			CompanyOffset: companyOffset,
			RemovalYear:   0xFFFF,
			RemovalMonth:  0xFF,
			RemovalDay:    0xFF,
			RatingID:      ratingID,
			IsNew:         0,
		})
	}

	l.Header.NumberOfDemoTables = uint32(len(l.DemoTable))
}

var demoSuffix = regexp.MustCompile(`(?i)(\s*[-:&]?\s*\b(demo|trial)\b.*|\s*\([^)]*\)|\s*-\s*(anagrams|crosswords|wordsearch).*|\s*theater\s*|\s*(cleaning|gift fossil|lasagna|mac & cheese|pots de crème|roasted red potatoes).*)$`)

func normalizeTitle(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return unicode.ToLower(r)
		}
		return -1
	}, s)
}

func MatchDemos(region constants.Region, gameTitle string) []info.DemoTable {
	want := normalizeTitle(gameTitle)
	if want == "" {
		return nil
	}

	var matches []info.DemoTable
	for _, demo := range RegionalDemos[region] {
		combined := strings.TrimSpace(demo.Title + " " + demo.Subtitle)
		stripped := demoSuffix.ReplaceAllString(combined, "")
		baseStripped := demoSuffix.ReplaceAllString(strings.TrimSpace(demo.Title), "")

		if normalizeTitle(combined) != want &&
			normalizeTitle(stripped) != want &&
			normalizeTitle(demo.Title) != want &&
			normalizeTitle(baseStripped) != want {
			continue
		}

		var entry info.DemoTable
		entry.ID = demo.ID
		copy(entry.Title[:], utf16.Encode([]rune(demo.Title)))
		copy(entry.Subtitle[:], utf16.Encode([]rune(demo.Subtitle)))
		matches = append(matches, entry)
	}

	return matches
}
