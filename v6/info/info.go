package info

import (
	"NintendoChannel/common"
	"NintendoChannel/constants"
	"NintendoChannel/gametdb"
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"os"
	"strings"
	"unicode/utf16"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mitchellh/go-wordwrap"
)

type Info struct {
	Header               Header
	SupportedControllers SupportedControllers
	SupportedFeatures    SupportedFeatures
	SupportedLanguages   SupportedLanguages
	_                    [10]byte
	Title                [31]uint16
	Subtitle             [31]uint16
	ShortTitle           [31]uint16
	DescriptionText      [3][41]uint16
	GenreText            [29]uint16
	PlayersText          [41]uint16
	PeripheralsText      [44]uint16
	_                    [80]byte
	DisclaimerText       [2400]uint16
	RatingID             uint8
	DistributionDateText [41]uint16
	WiiPointsText        [41]uint16
	CustomText           [10][41]uint16
	// Writing a blank one if it doesn't exist and not point to it
	// is way more efficient than writing everything individually
	TimePlayed          TimePlayed
	RecommendationTable RecommendationTable
}

var timePlayed = map[string]TimePlayed{}

func formatGenreText(rawGenre string) string {
	if strings.TrimSpace(rawGenre) == "" {
		return ""
	}

	displayNames := map[string]string{
		"2D platformer":        "2D Platformer",
		"3D platformer":        "3D Platformer",
		"first-person shooter": "First-Person Shooter",
		"third-person shooter": "Third-Person Shooter",
		"shoot 'em up":         "Shoot 'Em Up",
		"point-and-click":      "Point-and-Click",
		"role-playing":         "Role-Playing",
	}

	var genres []string
	for _, part := range strings.Split(rawGenre, ",") {
		key := strings.TrimSpace(part)
		if key == "" {
			continue
		}
		name, ok := displayNames[key]
		if !ok {
			words := strings.Fields(key)
			for idx, w := range words {
				if len(w) > 0 {
					words[idx] = strings.ToUpper(w[:1]) + w[1:]
				}
			}
			name = strings.Join(words, " ")
		}

		candidate := name
		if len(genres) > 0 {
			candidate = strings.Join(append(genres, name), ", ")
		}
		if len([]rune(candidate)) > 28 {
			break
		}
		genres = append(genres, name)
		if len(genres) == 3 {
			break
		}
	}

	return strings.Join(genres, ", ")
}

func (i *Info) MakeInfo(fileID uint32, game *gametdb.Game, title, synopsis string, region constants.Region, language constants.Language, titleType constants.TitleType, ratingDescriptors []string, recommendations map[string]common.TitleRecommendation, demos []DemoTable) {
	i.GetSupportedControllers(&game.Controllers)
	i.GetSupportedFeatures(game)
	i.GetSupportedLanguages(game.Languages)

	// Wrap title naturally across the two 30-character lines without stripping ':' or '-'
	if len([]rune(title)) > 30 {
		wrappedTitle := wordwrap.WrapString(title, 30)
		for p, line := range strings.Split(wrappedTitle, "\n") {
			if p == 0 {
				copy(i.Title[:30], utf16.Encode([]rune(line)))
			} else if p == 1 {
				copy(i.Subtitle[:30], utf16.Encode([]rune(line)))
			} else {
				break
			}
		}
	} else {
		copy(i.Title[:30], utf16.Encode([]rune(title)))
	}

	// Make tagline (first paragraph, 1-3 lines) and main synopsis (remaining text, up to 10 lines)
	paragraphs := strings.SplitN(strings.TrimSpace(strings.ReplaceAll(synopsis, "\r\n", "\n")), "\n", 2)
	taglineLines := strings.Split(wordwrap.WrapString(strings.TrimSpace(paragraphs[0]), 40), "\n")
	for i2, line := range taglineLines {
		if i2 >= 3 {
			break
		}
		copy(i.DescriptionText[i2][:], utf16.Encode([]rune(line)))
	}

	var bodyParts []string
	if len(taglineLines) > 3 {
		bodyParts = append(bodyParts, strings.Join(taglineLines[3:], " "))
	}
	if len(paragraphs) > 1 && strings.TrimSpace(paragraphs[1]) != "" {
		bodyParts = append(bodyParts, strings.TrimSpace(paragraphs[1]))
	}

	if len(bodyParts) > 0 {
		wrappedBody := strings.Split(wordwrap.WrapString(strings.Join(bodyParts, "\n"), 40), "\n")
		for i2, line := range wrappedBody {
			if i2 == 10 {
				break
			} else if i2 == 9 && strings.Contains(line, ".") {
				line = strings.Split(line, ".")[0] + "."
			}
			copy(i.CustomText[i2][:], utf16.Encode([]rune(line)))
		}
	}

	if genreText := formatGenreText(game.Genre); genreText != "" {
		copy(i.GenreText[:28], utf16.Encode([]rune(genreText)))
	}

	// Write the online players text if any
	if game.Features.OnlinePlayers != 0 {
		temp := []uint16{0, 0}
		copy(i.PlayersText[:], append(temp, utf16.Encode([]rune(fmt.Sprintf("%d Players (Online)", game.Features.OnlinePlayers)))...))
	}

	copy(i.DisclaimerText[:], utf16.Encode([]rune("Game information is provided by GameTDB.")))

	if v, ok := timePlayed[game.ID[:4]]; ok {
		i.Header.TimesPlayedTableOffset = 6744
		i.TimePlayed = v
	}

	if v, ok := recommendations[game.ID[:4]]; ok {
		i.Header.RatingTableOffset = 6744 + 16
		i.MakeRecommendationTable(v)
	}

	temp := new(bytes.Buffer)
	imageBuffer := new(bytes.Buffer)
	i.WriteAll(temp, imageBuffer)

	i.Header.PictureOffset = i.GetCurrentSize(imageBuffer)
	i.WriteCoverArt(imageBuffer, titleType, region, game.ID)
	i.WriteRatingImage(imageBuffer, region)
	i.WriteRatingDescriptor(imageBuffer, region, ratingDescriptors)
	i.WriteDemos(imageBuffer, demos)
	i.Header.Filesize = i.GetCurrentSize(imageBuffer)
	temp.Reset()

	i.WriteAll(temp, imageBuffer)
	crcTable := crc32.MakeTable(crc32.IEEE)
	checksum := crc32.Checksum(temp.Bytes(), crcTable)
	i.Header.CRC32 = checksum
	temp.Reset()

	i.WriteAll(temp, imageBuffer)

	// Ensure write path exists
	config := common.GetConfig()
	err := os.MkdirAll(fmt.Sprintf("%s/infos/%d/%d", config.AssetsPath, region, language), 0777)
	common.CheckError(err)

	err = os.WriteFile(fmt.Sprintf("%s/infos/%d/%d/%d.info", config.AssetsPath, region, language, fileID), temp.Bytes(), 0666)
	common.CheckError(err)
}

func (i *Info) WriteAll(buffer, imageBuffer *bytes.Buffer) {
	err := binary.Write(buffer, binary.BigEndian, *i)
	common.CheckError(err)

	buffer.Write(imageBuffer.Bytes())
}

func (i *Info) GetCurrentSize(_buffer *bytes.Buffer) uint32 {
	buffer := bytes.NewBuffer(nil)
	i.WriteAll(buffer, _buffer)
	return uint32(buffer.Len())
}

func GetTimePlayed(ctx *context.Context, pool *pgxpool.Pool) {
	rows, err := pool.Query(*ctx, `SELECT game_id, COUNT(game_id), SUM(times_played), SUM(time_played) FROM time_played GROUP BY game_id`)
	common.CheckError(err)

	for rows.Next() {
		var gameID string
		var numberOfPlayers int
		var totalTimesPlayed int
		var totalTimePlayed int

		err = rows.Scan(&gameID, &numberOfPlayers, &totalTimesPlayed, &totalTimePlayed)
		common.CheckError(err)

		timePlayed[gameID] = TimePlayed{
			TotalTimePlayed:           uint32(totalTimePlayed / 60),
			TimeSpentPlayingPerPerson: uint32(totalTimePlayed / numberOfPlayers),
			TotalTimesPlayed:          uint32(totalTimesPlayed),
			TimesPlayedPerPerson:      uint32((float64(totalTimesPlayed / numberOfPlayers)) / 0.01),
		}
	}
}
