package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

type Words map[string]int

// sum sums the values of the counter.
func (c Words) sum() int {
	var sum int
	for _, v := range c {
		sum += v
	}
	return sum
}

// readWords reads a file and returns a slice of all the words in the file.
func readWords(name string) []string {
	file, err := os.Open(name)
	if err != nil {
		log.Fatalln(err)
	}
	r := regexp.MustCompile("\\w+")
	var words []string

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		w := r.FindString(scanner.Text())
		words = append(words, strings.ToLower(w))
	}
	if err := scanner.Err(); err != nil {
		log.Fatalln("invalid input", err)
	}
	return words
}

// countWords counts all the items in the words slice, words are stored as keys
// and their counts are stored as values.
func countWords(words []string) Words {
	c := make(Words)
	for _, w := range words {
		c[w] += 1
	}
	return c
}

// p calculates the probability of word.
func p(word string, words Words) float64 {
	n := words.sum()
	return float64(words[word]) / float64(n)
}

// correction most probable spelling correction for word.
func correction(word string, allWords Words) string {
	return max(candidates(word, allWords), allWords)
}

// candidates generate possible spelling corrections for word.
func candidates(word string, allWords Words) []string {
	c := []string{word}

	k := known(c, allWords)
	if len(k) > 0 {
		return k
	}

	k = known(edits1(word), allWords)
	if len(k) > 0 {
		return k
	}

	k = known(edits2(word), allWords)
	if len(k) > 0 {
		return k
	}

	return c
}

// known returns a subset of 'words' if they appear in 'allWords'.
func known(words []string, allWords Words) []string {
	var items []string
	for _, w := range words {
		if _, ok := allWords[w]; ok {
			items = append(items, w)
		}
	}
	return items
}

// max chooses the candidate with the highest combined probability.
func max(words []string, allWords Words) string {
	var m float64 = -1
	var s string
	for _, w := range words {
		t := p(w, allWords)
		if t > m {
			m = t
			s = w
		}
	}
	return s
}

// edits1 all edits that are one edit away from word.
func edits1(word string) []string {
	var splits []string
	for i := 0; i < len(word)+1; i++ {
		splits = append(splits, word[:i])
		splits = append(splits, word[i:])
	}

	var deletes []string
	for i := 0; i < len(word); i++ {
		deletes = append(deletes, word[:i]+word[i+1:])
	}

	var transposes []string
	for i := 0; i < len(splits)-1; i++ {
		l, r := splits[i], splits[i+1]
		if len(r) > 1 {
			q := l + string(r[1]) + string(r[0]) + r[2:]
			if len(q) == len(word) {
				transposes = append(transposes, q)
			}
		}
	}

	const letters = "abcdefghijklmnopqrstuvwxyz"

	var replaces []string
	for i := 0; i < len(splits)-1; i += 2 {
		l, r := splits[i], splits[i+1]
		if len(r) > 0 {
			for _, c := range letters {
				replaces = append(replaces, l+string(c)+r[1:])
			}
		}
	}

	var inserts []string
	for i := 0; i < len(splits)-1; i += 2 {
		l, r := splits[i], splits[i+1]
		for _, c := range letters {
			inserts = append(inserts, l+string(c)+r)
		}
	}

	all := affix(deletes, transposes, replaces, inserts)

	// remove duplicates
	set := make(map[string]struct{})
	for _, item := range all {
		set[item] = struct{}{}
	}

	var edits []string
	for k, _ := range set {
		edits = append(edits, k)
	}

	return edits
}

// edits2 all edits that are two edits away from a given word.
func edits2(word string) []string {
	var edits []string
	for _, e1 := range edits1(word) {
		for _, e2 := range edits1(e1) {
			edits = append(edits, e2)
		}
	}
	return edits
}

// affix creates a single slice from multiple slices.
func affix(items ...[]string) []string {
	var all []string
	for _, i := range items {
		all = append(all, i...)
	}
	return all
}

func main() {
	words := readWords("big.txt")
	count := countWords(words)

	if len(os.Args) == 1 {
		fmt.Println("usage: spelling-corrector <mis-spelt word> e.g. spelling-corrector peotryy")
		return
	}

	w := os.Args[1]
	fmt.Println(correction(w, count))
}
