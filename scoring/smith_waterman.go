package scoring

import (
	"fmt"
	"os"
	"unicode"
)

// smithWaterman calculates a simularity score between s1 and s2
// by smith-waterman algorithm. smith-waterman algorithm is one of
// local alignment algorithms and it uses dynamic programming.
//
// In this smith-waterman algorithm, we use the affine gap penalty.
// Please see https://en.wikipedia.org/wiki/Gap_penalty#Affine for additional
// information about the affine gap penalty.
//
// We calculate the gap penalty by the Gotoh's algorithm, which optimizes
// the calculation from O(M^2N) to O(MN).
// Please see ftp://150.128.97.71/pub/Bioinformatica/gotoh1982.pdf for more details.
func smithWaterman(s1, s2 []rune) int {
	const (
		openGap int32 = 5 // Gap opening penalty.
		extGap  int32 = 1 // Gap extension penalty.

		matchScore    int32 = 5
		mismatchScore int32 = 1

		firstCharBonus int32 = 3 // The first char of s1 is equal to s2's one.
	)

	// The scoring matrix.
	H := make([][]int32, len(s1)+1)
	// A matrix that calculates gap penalties for s2 until each position (i, j).
	// Note that, we don't need a matrix for s1 because s1 contains all runes
	// of s2 so that s1 is not inserted gaps.
	D := make([][]int32, len(s1)+1)
	for i := 0; i <= len(s1); i++ {
		H[i] = make([]int32, len(s2)+1)
		D[i] = make([]int32, len(s2)+1)
	}

	for i := 0; i <= len(s1); i++ {
		D[i][0] = -openGap - int32(i)*extGap
	}

	// Calculate bonuses for each rune of s1.
	bonus := make([]int32, len(s1))
	bonus[0] = firstCharBonus
	prevCh := s1[0]
	prevIsDelimiter := isDelimiter(prevCh)
	for i, r := range s1[1:] {
		isDelimiter := isDelimiter(r)
		if prevIsDelimiter && !isDelimiter {
			bonus[i] = firstCharBonus
		}
		prevIsDelimiter = isDelimiter
	}

	var maxScore int32
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			var score int32
			if s1[i-1] != s2[j-1] {
				score = H[i-1][j-1] - mismatchScore
			} else {
				score = H[i-1][j-1] + matchScore + bonus[i-1]
			}
			H[i][j] += max(D[i-1][j], score, 0)

			D[i][j] = max(H[i-1][j]-openGap, D[i-1][j]-extGap)

			// Update the max score.
			maxScore = max(H[i][j], maxScore)
		}
	}

	if isDebug() {
		printSlice := func(m [][]int32) {
			fmt.Printf("%4c     ", '|')
			for i := 0; i < len(s2); i++ {
				fmt.Printf("%3c ", s2[i])
			}
			fmt.Printf("\n-------------------------\n")

			fmt.Print("   | ")
			for i := 0; i <= len(s1); i++ {
				if i != 0 {
					fmt.Printf("%3c| ", s1[i-1])
				}
				for j := 0; j <= len(s2); j++ {
					fmt.Printf("%3d ", m[i][j])
				}
				fmt.Println()
			}
			println()
		}
		printSlice(H)
		printSlice(D)
	}

	// We adjust scores by the weight per one rune.
	return int(float32(maxScore) * (float32(maxScore) / float32(len(s1))))
}

func isDebug() bool {
	return os.Getenv("DEBUG") != ""
}

var delimiterRunes = map[rune]interface{}{
	'(': nil,
	'[': nil,
	'{': nil,
	'/': nil,
	'-': nil,
	'_': nil,
	'.': nil,
}

func isDelimiter(r rune) bool {
	if _, ok := delimiterRunes[r]; ok {
		return true
	}
	return unicode.IsSpace(r)
}