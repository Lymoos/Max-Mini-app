package main

import (
	"sort"
	"strings"
	"unicode"
)

type CatalogItem struct {
	ID      string
	Name    string
	Aliases []string
}

type Match struct {
	ID        string
	Name      string
	Matched   string
	Corrected bool
	dist      int
}

func normalize(s string) []rune {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "ё", "е")
	result := []rune{}
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result = append(result, r)
		}
	}
	return result
}

var qwertyToRu = map[rune]rune{
	'q': 'й', 'w': 'ц', 'e': 'у', 'r': 'к', 't': 'е', 'y': 'н', 'u': 'г', 'i': 'ш', 'o': 'щ', 'p': 'з',
	'[': 'х', ']': 'ъ', 'a': 'ф', 's': 'ы', 'd': 'в', 'f': 'а', 'g': 'п', 'h': 'р', 'j': 'о', 'k': 'л',
	'l': 'д', ';': 'ж', '\'': 'э', 'z': 'я', 'x': 'ч', 'c': 'с', 'v': 'м', 'b': 'и', 'n': 'т', 'm': 'ь',
	',': 'б', '.': 'ю', '`': 'е',
}

// текст, набранный в английской раскладке, переводим в русскую: gfhfwtnfvjk -> парацетамол
func fixLayout(s string) (string, bool) {
	changed := false
	result := []rune{}
	for _, r := range strings.ToLower(s) {
		if ru, ok := qwertyToRu[r]; ok {
			result = append(result, ru)
			changed = true
		} else {
			result = append(result, r)
		}
	}
	return string(result), changed
}

func levenshtein(a, b []rune) int {
	return editDistance(a, b, false)
}

// то же расстояние, но путаница безударных гласных (малако — молоко) не считается ошибкой
func typoDistance(a, b []rune) int {
	return editDistance(a, b, true)
}

var similarVowels = map[[2]rune]bool{
	{'а', 'о'}: true, {'о', 'а'}: true,
	{'е', 'и'}: true, {'и', 'е'}: true,
	{'и', 'ы'}: true, {'ы', 'и'}: true,
	{'е', 'э'}: true, {'э', 'е'}: true,
}

// соседние переставленные буквы (шрфит — шрифт) считаем одной ошибкой
func editDistance(a, b []rune, vowelsFree bool) int {
	prev2 := make([]int, len(b)+1)
	prev := make([]int, len(b)+1)
	cur := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] || (vowelsFree && similarVowels[[2]rune{a[i-1], b[j-1]}]) {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
			if i > 1 && j > 1 && a[i-1] == b[j-2] && a[i-2] == b[j-1] {
				cur[j] = min(cur[j], prev2[j-2]+1)
			}
		}
		prev2, prev, cur = prev, cur, prev2
	}
	return prev[len(b)]
}

func maxTypos(length int) int {
	switch {
	case length < 4:
		return 0
	case length <= 5:
		return 1
	case length <= 9:
		return 2
	default:
		return 3
	}
}

// во фразе слов много, поэтому ошибок прощаем меньше, чем в подсказках: иначе «привет» станет «Париет»
func textTypos(length int) int {
	switch {
	case length <= 5:
		return 0
	case length <= 8:
		return 1
	default:
		return 2
	}
}

// сравниваем запрос с началом названия, чтобы подсказки работали во время набора
func prefixDistance(query, name []rune) int {
	if len(query) <= len(name) && string(name[:len(query)]) == string(query) {
		return 0
	}
	d := typoDistance(query, name)
	if len(query) < len(name) {
		d = min(d, typoDistance(query, name[:len(query)]))
	}
	return d
}

func fuzzySuggest(items []CatalogItem, query string, limit int) []Match {
	queries := [][]rune{normalize(query)}
	fixed, changed := fixLayout(query)
	if changed {
		queries = append(queries, normalize(fixed))
	}

	result := []Match{}
	for _, item := range items {
		best := Match{dist: -1}
		names := append([]string{item.Name}, item.Aliases...)
		for i, q := range queries {
			if len(q) < 2 {
				continue
			}
			for _, name := range names {
				n := normalize(name)
				d := prefixDistance(q, n)
				if d > maxTypos(len(q)) {
					continue
				}
				exact := len(q) <= len(n) && string(n[:len(q)]) == string(q)
				if best.dist == -1 || d < best.dist {
					best = Match{ID: item.ID, Name: item.Name, Matched: name, Corrected: !exact || i > 0, dist: d}
				}
			}
		}
		if best.dist != -1 {
			result = append(result, best)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].dist != result[j].dist {
			return result[i].dist < result[j].dist
		}
		return result[i].Name < result[j].Name
	})
	if len(result) > limit {
		result = result[:limit]
	}
	return result
}

// слова, которыми называют целую группу лекарств: «витамины» — это не «Витамин C»
var genericWords = []string{"витамин", "таблет", "лекарств", "капл", "мазь", "мази", "крем", "гель", "сироп", "спрей", "средств", "препарат"}

func isGeneric(word []rune) bool {
	for _, g := range genericWords {
		if strings.HasPrefix(string(word), g) {
			return true
		}
	}
	return false
}

// ищем название из каталога среди слов фразы: «где купить нурофена» -> Ибупрофен.
// Названия бывают из нескольких слов, поэтому проверяем и сочетания до трёх слов подряд
func findInText(items []CatalogItem, text string) (CatalogItem, bool) {
	words := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-'
	})
	// сочетания, которые начинаются с общего слова («витамин d3»), принимаем только без ошибок
	type candidate struct {
		word   []rune
		strict bool
	}
	candidates := []candidate{}
	for i := range words {
		joined := []rune{}
		generic := false
		for j := i; j < len(words) && j < i+3; j++ {
			joined = append(joined, normalize(words[j])...)
			if j == i {
				generic = isGeneric(joined)
				if generic {
					continue
				}
			}
			candidates = append(candidates, candidate{append([]rune{}, joined...), generic})
		}
	}

	bestDist := -1
	var best CatalogItem
	for _, c := range candidates {
		word := c.word
		if len(word) < 4 {
			continue
		}
		for _, item := range items {
			for _, name := range append([]string{item.Name}, item.Aliases...) {
				n := normalize(name)
				if len(n) < 4 || len(word) < len(n)-2 {
					continue
				}
				if c.strict {
					if string(word) == string(n) {
						return item, true
					}
					continue
				}
				d := typoDistance(word, n)
				if len(word) > len(n) && len(word)-len(n) <= 3 {
					d = min(d, typoDistance(word[:len(n)], n))
				}
				if d <= textTypos(len(n)) && (bestDist == -1 || d < bestDist) {
					bestDist = d
					best = item
				}
			}
		}
	}
	return best, bestDist != -1
}
