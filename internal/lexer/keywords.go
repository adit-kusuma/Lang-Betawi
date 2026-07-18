package lexer

var KeywordMap = map[TokenType][]string{
	PRINT: {
		"ngomong", "kasiTau", "nyablak", "cuap",
		"ngecap", "cerocos", "ngoceh", "cakcek",
	},
	ASSIGN: {
		"entu", "tuh",
	},
	IF: {
		"kalo", "kalu",
	},
	ELSE: {
		"kaloKagak", "kaloKaga", "laenKagak",
	},
	LOOP: {
		"musing", "ngulang", "terosNgulang",
	},
	FUNCTION: {
		"bikinGaya", "bikinJurus", "bikinAksi",
	},
	IMPORT: {
		"bawa", "ngajak", "colek",
	},
	RETURN: {
		"balikin", "kasihBalik", "baliqin",
	},
	TRUE: {
		"bener", "beneran", "topMarkotop",
	},
	FALSE: {
		"kagak", "kaga", "kagaDah", "gakada",
	},
	NULL_LIT: {
		"zonk", "kapiran", "keder",
	},
	SERVER_START: {
		"bukaWarung", "bukaLapakGede",
	},
	ROUTE_DEF: {
		"bikinLapak", "gelarLapak",
	},
	DB_QUERY: {
		"tanyaDatabase", "nanyaKeDatabase", "colekDatabase", "cariinData",
	},
}

var reverseKeywords map[string]TokenType

type fuzzyCandidate struct {
	Word string
	Type TokenType
}

var fuzzyCandidates []fuzzyCandidate

func init() {
	buildIndexes()
}

func buildIndexes() {
	reverseKeywords = make(map[string]TokenType)
	fuzzyCandidates = nil

	for tokType, words := range KeywordMap {
		for _, w := range words {
			reverseKeywords[w] = tokType
			fuzzyCandidates = append(fuzzyCandidates, fuzzyCandidate{
				Word: w,
				Type: tokType,
			})
		}
	}
}

func RegisterDialectPack(pack map[TokenType][]string) {
	for tokType, words := range pack {
		KeywordMap[tokType] = append(KeywordMap[tokType], words...)
	}
	buildIndexes()
}

func LookupExact(word string) (TokenType, bool) {
	tok, ok := reverseKeywords[word]
	return tok, ok
}
