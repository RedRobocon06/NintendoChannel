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
	Unknown       uint32
	RatingID      uint8
	IsNew         uint8
	NewTagIndex   uint8
	_             [205]byte
}

type DemoEntry struct {
	ID            uint32
	Title         string
	Subtitle      string
	TitleID       uint32
	CompanyOffset uint32
	RemovalYear   uint16
	RemovalMonth  uint8
	RemovalDay    uint8
	Unknown       uint32
	RatingID      uint8
	IsNew         uint8
	NewTagIndex   uint8
}

var RegionalDemos = map[constants.Region][]DemoEntry{
	constants.NTSC: {
		{ID: 839603576, Title: "100 Classic Books", TitleID: 1737258670, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 76014458, Title: "America's Test Kitchen", Subtitle: "Pots de Crème Demo", TitleID: 1764120452, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 941060636, Title: "America's Test Kitchen", Subtitle: "Roasted Red Potatoes Demo", TitleID: 1764120452, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 839603577, Title: "Art Academy", TitleID: 1737258670, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 839603578, Title: "Big Brain Academy", TitleID: 1737258670, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 359174269, Title: "Brain Age Demo", TitleID: 33884163, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 568827099, Title: "Brain Age 2 Demo", TitleID: 284234639, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 839603579, Title: "Clubhouse Games", Subtitle: "(Darts)", TitleID: 1737258670, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 839603580, Title: "Clubhouse Games", Subtitle: "(Turncoat)", TitleID: 1737258670, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 839603581, Title: "Crosswords DS-Anagrams Demo", TitleID: 1737258670, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 152017202, Title: "Crosswords DS-Crosswords Demo", TitleID: 1599282847, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 839603582, Title: "Crosswords DS-Wordsearch Demo", TitleID: 1737258670, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 839603583, Title: "Custom Robo Arena", TitleID: 1737258670, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 10},
		{ID: 839603585, Title: "Elite Beat Agents", TitleID: 1737258670, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 10},
		{ID: 1972710029, Title: "Flash Focus Demo", TitleID: 697480931, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 839603586, Title: "Flipnote Studio Theater", TitleID: 1737258670, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 1141533641, Title: "Fossil Fighters", Subtitle: "Cleaning Demo", TitleID: 1737258670, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 1687629128, Title: "Fossil Fighters", Subtitle: "Gift Fossil (Air)", TitleID: 1737258670, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 548229158, Title: "Fossil Fighters", Subtitle: "Gift Fossil (Earth)", TitleID: 1737258670, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 918716973, Title: "Fossil Fighters", Subtitle: "Gift Fossil (Fire)", TitleID: 1737258670, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 262123039, Title: "Fossil Fighters", Subtitle: "Gift Fossil (Neutral)", TitleID: 1737258670, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 1557878207, Title: "Fossil Fighters", Subtitle: "Gift Fossil (Water)", TitleID: 1737258670, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 2144771072, Title: "Fossil Fighters: Champions", TitleID: 1206858045, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 2043275873, Title: "Kirby Mass Attack", TitleID: 302350231, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 839603587, Title: "Magnetica", TitleID: 1737258670, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 839603588, Title: "Mario Kart DS", TitleID: 1737258670, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 839603589, Title: "Mario vs. Donkey Kong 2:", Subtitle: "March of the Minis", TitleID: 1737258670, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 839603590, Title: "Mario vs. Donkey Kong:", Subtitle: "Mini-Land Mayhem!", TitleID: 1737258670, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 839603591, Title: "Meteos", TitleID: 1737258670, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 839603592, Title: "Metroid Prime Hunters", TitleID: 1737258670, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 11},
		{ID: 270763606, Title: "Mystery Case Files:", Subtitle: "MillionHeir", TitleID: 2042077528, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 1879850935, Title: "Personal Trainer: Cooking", Subtitle: "Lasagna Demo", TitleID: 1737258670, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 187959385, Title: "Personal Trainer: Cooking", Subtitle: "Mac & Cheese Demo", TitleID: 1737258670, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 692916852, Title: "Personal Trainer: Math", TitleID: 923223269, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 340771894, Title: "Picross 3D", TitleID: 1652375569, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 839603593, Title: "Picross DS", TitleID: 1737258670, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 839603594, Title: "Planet Puzzle League", TitleID: 1737258670, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 839603595, Title: "Pokémon Mystery Dungeon:", Subtitle: "Explorers of Sky Demo", TitleID: 1737258670, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 2147201832, Title: "Pokémon Ranger: Guardian Signs", Subtitle: "Demo", TitleID: 2066395031, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 839603596, Title: "Pokémon Trozei!", TitleID: 1737258670, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 1308838811, Title: "Rhythm Heaven", TitleID: 1231101988, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 839603597, Title: "Tetris DS", Subtitle: "(Push)", TitleID: 1737258670, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 839603598, Title: "Tetris DS", Subtitle: "(Standard)", TitleID: 1737258670, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
		{ID: 839603599, Title: "True Swing Golf", TitleID: 1737258670, CompanyOffset: 27296, RemovalYear: 65535, RemovalMonth: 255, RemovalDay: 255, RatingID: 9},
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
		if demo.TitleID != 0 && t.ID == demo.TitleID {
			return t
		}
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

		titleID := demo.TitleID
		ratingID := demo.RatingID
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
			RemovalYear:   demo.RemovalYear,
			RemovalMonth:  demo.RemovalMonth,
			RemovalDay:    demo.RemovalDay,
			Unknown:       demo.Unknown,
			RatingID:      ratingID,
			IsNew:         demo.IsNew,
			NewTagIndex:   demo.NewTagIndex,
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

func MatchDemos(region constants.Region, titleID uint32, gameTitle string) []info.DemoTable {
	want := normalizeTitle(gameTitle)
	if want == "" && titleID == 0 {
		return nil
	}

	var matches []info.DemoTable
	for _, demo := range RegionalDemos[region] {
		combined := strings.TrimSpace(demo.Title + " " + demo.Subtitle)
		stripped := demoSuffix.ReplaceAllString(combined, "")
		baseStripped := demoSuffix.ReplaceAllString(strings.TrimSpace(demo.Title), "")

		idMatch := titleID != 0 && demo.TitleID == titleID
		nameMatch := want != "" && (normalizeTitle(combined) == want || normalizeTitle(stripped) == want || normalizeTitle(demo.Title) == want || normalizeTitle(baseStripped) == want)
		if !idMatch && !nameMatch {
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
