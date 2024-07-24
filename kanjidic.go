package yomitan

import (
	"os"
	"strconv"

	jmdict "github.com/themoeway/jmdict-go"
)

func kanjidicExtractKanji(entry jmdict.KanjidicCharacter, language string) *dbKanji {
	if entry.ReadingMeaning == nil {
		return nil
	}

	kanji := dbKanji{
		Character: entry.Literal,
		Stats:     make(map[string]string),
	}

	for _, m := range entry.ReadingMeaning.Meanings {
		if m.Language == nil && language == "" || m.Language != nil && language == *m.Language {
			kanji.Meanings = append(kanji.Meanings, m.Meaning)
		}
	}

	if len(kanji.Meanings) == 0 {
		return nil
	}

	for _, number := range entry.DictionaryNumbers {
		kanji.Stats[number.Type] = number.Value
	}

	if frequency := entry.Misc.Frequency; frequency != nil {
		kanji.Stats["freq"] = *frequency
	}

	if level := entry.Misc.JlptLevel; level != nil {
		kanji.Stats["jlpt"] = *level
	}

	if counts := entry.Misc.StrokeCounts; len(counts) > 0 {
		kanji.Stats["strokes"] = counts[0]
	}

	for _, code := range entry.Codepoint {
		kanji.Stats[code.Type] = code.Value
	}

	for _, code := range entry.QueryCode {
		kanji.Stats[code.Type] = code.Value
	}

	if grade := entry.Misc.Grade; grade != nil {
		kanji.Stats["grade"] = *grade
		if gradeInt, err := strconv.Atoi(*grade); err == nil {
			if gradeInt >= 1 && gradeInt <= 8 {
				kanji.addTags("jouyou")
			} else if gradeInt >= 9 && gradeInt <= 10 {
				kanji.addTags("jinmeiyou")
			}
		}
	}

	for _, r := range entry.ReadingMeaning.Readings {
		switch r.Type {
		case "ja_on":
			kanji.Onyomi = append(kanji.Onyomi, r.Value)
		case "ja_kun":
			kanji.Kunyomi = append(kanji.Kunyomi, r.Value)
		}
	}

	return &kanji
}

func kanjidicExportDb(inputPath, outputPath, language, title string, stride int, pretty bool) error {
	reader, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	dict, err := jmdict.LoadKanjidic(reader)
	if err != nil {
		return err
	}

	var langTag string
	switch language {
	case "french":
		langTag = "fr"
	case "spanish":
		langTag = "es"
	case "portuguese":
		langTag = "pt"
	}

	var kanji dbKanjiList
	for _, entry := range dict.Characters {
		kanjiCurr := kanjidicExtractKanji(entry, langTag)
		if kanjiCurr != nil {
			kanji = append(kanji, *kanjiCurr)
		}
	}

	if title == "" {
		title = "KANJIDIC2"
	}
	kanjidicDate := dict.Header.DatabaseVersion
	title = title + " [" + kanjidicDate + "]"

	tags := dbTagList{
		dbTag{Name: "jouyou", Notes: "included in list of regular-use characters", Category: "frequent", Order: -5},
		dbTag{Name: "jinmeiyou", Notes: "included in list of characters for use in personal names", Category: "frequent", Order: -5},

		dbTag{Name: "freq", Notes: "Frequency", Category: "misc"},
		dbTag{Name: "grade", Notes: "Grade level", Category: "misc"},
		dbTag{Name: "jlpt", Notes: "JLPT level", Category: "misc"},
		dbTag{Name: "strokes", Notes: "Stroke count", Category: "misc"},

		dbTag{Name: "jis208", Notes: "JIS X 0208-1997 kuten code", Category: "code"},
		dbTag{Name: "jis212", Notes: "JIS X 0212-1990 kuten code", Category: "code"},
		dbTag{Name: "jis213", Notes: "JIS X 0213-2000 kuten code", Category: "code"},
		dbTag{Name: "ucs", Notes: "Unicode hex code", Category: "code"},

		dbTag{Name: "deroo", Notes: "2001 Kanji", Category: "class"},
		dbTag{Name: "four_corner", Notes: "Four corner code", Category: "class"},
		dbTag{Name: "misclass", Notes: "Misclassification", Category: "class"},
		dbTag{Name: "sh_desc", Notes: "The Kanji Dictionary", Category: "class"},
		dbTag{Name: "skip", Notes: "SKIP code", Category: "class"},

		dbTag{Name: "busy_people", Notes: "Japanese For Busy People", Category: "index"},
		dbTag{Name: "crowley", Notes: "The Kanji Way to Japanese Language Power", Category: "index"},
		dbTag{Name: "gakken", Notes: "A  New Dictionary of Kanji Usage", Category: "index"},
		dbTag{Name: "halpern_kkd", Notes: "Kodansha Kanji Dictionary", Category: "index"},
		dbTag{Name: "halpern_kkld", Notes: "Kanji Learners Dictionary", Category: "index"},
		dbTag{Name: "halpern_kkld_2ed", Notes: "Kanji Learners Dictionary", Category: "index"},
		dbTag{Name: "halpern_njecd", Notes: "New Japanese-English Character Dictionary", Category: "index"},
		dbTag{Name: "heisig", Notes: "Remembering The  Kanji", Category: "index"},
		dbTag{Name: "heisig6", Notes: "Remembering The  Kanji, Sixth Ed.", Category: "index"},
		dbTag{Name: "henshall", Notes: "A Guide To Remembering Japanese Characters", Category: "index"},
		dbTag{Name: "henshall3", Notes: "A Guide To Reading and Writing Japanese", Category: "index"},
		dbTag{Name: "jf_cards", Notes: "Japanese Kanji Flashcards", Category: "index"},
		dbTag{Name: "kanji_in_context", Notes: "Kanji in Context", Category: "index"},
		dbTag{Name: "kodansha_compact", Notes: "Kodansha Compact Kanji Guide", Category: "index"},
		dbTag{Name: "maniette", Notes: "Les Kanjis dans la tete", Category: "index"},
		dbTag{Name: "moro", Notes: "Daikanwajiten", Category: "index"},
		dbTag{Name: "nelson_c", Notes: "Modern Reader's Japanese-English Character Dictionary", Category: "index"},
		dbTag{Name: "nelson_n", Notes: "The New Nelson Japanese-English Character Dictionary", Category: "index"},
		dbTag{Name: "oneill_kk", Notes: "Essential Kanji", Category: "index"},
		dbTag{Name: "oneill_names", Notes: "Japanese Names", Category: "index"},
		dbTag{Name: "sakade", Notes: "A Guide To Reading and Writing Japanese", Category: "index"},
		dbTag{Name: "sh_kk", Notes: "Kanji and Kana", Category: "index"},
		dbTag{Name: "sh_kk2", Notes: "Kanji and Kana", Category: "index"},
		dbTag{Name: "tutt_cards", Notes: "Tuttle Kanji Cards", Category: "index"},
	}

	recordData := map[string]dbRecordList{
		"kanji": kanji.crush(),
		"tag":   tags.crush(),
	}

	index := dbIndex{
		Title:       title,
		Revision:    "kanjidic2." + kanjidicDate,
		Sequenced:   false,
		Attribution: edrdgAttribution,
	}

	return writeDb(
		outputPath,
		index,
		recordData,
		stride,
		pretty,
	)
}
