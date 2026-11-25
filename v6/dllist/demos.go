package dllist

import (
	"NintendoChannel/common"
	"NintendoChannel/constants"
	"NintendoChannel/v6/info"
	"bytes"
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

var demosTable []*NinchDllist_DemosTable

func PopulateDemos() {
	dl := NewNinchDllist()
	err := dl.Read(kaitai.NewStream(bytes.NewReader(constants.DLList)), nil, dl)
	common.CheckError(err)

	demosTable, err = dl.DemosTable()
	common.CheckError(err)
}

func decodeUTF16Z(u []uint16) string {
	n := 0
	for n < len(u) && u[n] != 0 {
		n++
	}
	return string(utf16.Decode(u[:n]))
}

func (l *List) findTitleForDemo(demo *NinchDllist_DemosTable) *TitleTable {
	stripped := demoSuffix.ReplaceAllString(strings.TrimSpace(demo.Title+" "+demo.Subtitle), "")
	wantFull := normalizeTitle(stripped)
	wantBase := normalizeTitle(demo.Title)

	var fallback *TitleTable
	for idx := range l.TitleTable {
		t := &l.TitleTable[idx]
		tName := strings.TrimSpace(decodeUTF16Z(t.TitleName[:]) + " " + decodeUTF16Z(t.Subtitle[:]))
		tBase := decodeUTF16Z(t.TitleName[:])
		nFull := normalizeTitle(tName)
		nBase := normalizeTitle(tBase)

		if (wantFull != "" && (nFull == wantFull || nBase == wantFull)) || (wantBase != "" && (nFull == wantBase || nBase == wantBase)) {
			if t.TitleType == constants.NintendoDS {
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

	for _, demo := range demosTable {
		var title [31]uint16
		tempTitle := utf16.Encode([]rune(demo.Title))
		copy(title[:], tempTitle)

		var subtitle [31]uint16
		tempSubtitle := utf16.Encode([]rune(demo.Subtitle))
		copy(subtitle[:], tempSubtitle)

		titleID := demo.Titleid
		ratingID := demo.RatingId
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
                        ID:            demo.Id,
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

var demoSuffix = regexp.MustCompile(`(?i)\s*[-:&]*\s*\b(demo|trial)\b.*`)

func normalizeTitle(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return unicode.ToLower(r)
		}
		return -1
	}, s)
}

func MatchDemos(gameTitle string) []info.DemoTable {
	want := normalizeTitle(gameTitle)
	if want == "" {
		return nil
	}

	var matches []info.DemoTable
	for _, demo := range demosTable {
		name := demoSuffix.ReplaceAllString(strings.TrimSpace(demo.Title+" "+demo.Subtitle), "")
		if normalizeTitle(name) != want && normalizeTitle(demo.Title) != want {
			continue
		}

		var entry info.DemoTable
		entry.ID = demo.Id
		copy(entry.Title[:], utf16.Encode([]rune(demo.Title)))
		copy(entry.Subtitle[:], utf16.Encode([]rune(demo.Subtitle)))
		matches = append(matches, entry)
	}

	return matches
}
