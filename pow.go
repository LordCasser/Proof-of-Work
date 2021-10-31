package main

import (
	// "github.com/minio/sha256-simd"
	"crypto/sha256"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

var resultChan = make(chan Result)
var idleChan chan int

type Result struct {
	attempts  int
	binString string
}

// Solver is a pow solver
type Solver struct {
	SolverConfig
}

// SolverConfig is config struct for Solver
type SolverConfig struct {
	Prefix     string
	Difficulty int
}

// New returns a POW solver
func New(c SolverConfig) (*Solver, error) {
	if c.Difficulty <= 0 {
		return nil, errors.New("difficulty should > 0")
	}
	return &Solver{
		SolverConfig{
			Prefix:     c.Prefix,
			Difficulty: c.Difficulty,
		},
	}, nil
}

//const k = 1000000
const k = 10000000

// Solve returns two fields,
// @ret1 the number of iteration after solving the pow.
// @ret2 binary form of the answer.
func (s *Solver) Solve(start int) {
	var attempts = start * k
	var binaryString string
	for attempts < (start+1)*k {
		sum := sha256.Sum256([]byte(fmt.Sprintf("%s%v", s.Prefix, attempts)))
		if isBinValid(sum, s.Difficulty) {
			binaryString = toBinString(sum)
			if !isValid(binaryString, s.Difficulty) {
				attempts++
				continue
			}
			resultChan <- Result{attempts: attempts, binString: binaryString}
			break
		}
		attempts++
	}
	defer func() {
		idleChan <- 1
		fmt.Println("[@]" + strconv.Itoa(attempts) + " Done...")
	}()
}

func toBinString(s [sha256.Size]byte) string {
	var ss string
	for _, b := range s {
		ss = fmt.Sprintf("%s%08b", ss, b)
	}
	return ss
}

func isBinValid(s [sha256.Size]byte, d int) bool {
	for i := 0; i < d/8; i++ {
		if s[i] != 0 {
			return false
		}
	}
	return (byte((1<<(8-(d-(d/8)*8)))-1)^0xff)&s[d/8] == 0
}

func isValid(ss string, d int) bool {
	for i := range ss {
		if i >= d {
			break
		}
		if ss[i] != '0' {
			return false
		}
	}
	return true
}
func main() {
	var prefix = flag.String("p", "", "prefix")
	var difficulty = flag.Int("d", 26, "difficulty")
	var binString = flag.Bool("b", false, "output binString or not")
	var threshold = flag.Int("t", 10, "threshold")
	flag.Parse()
	if *prefix == "" {
		flag.Usage()
		os.Exit(1)
	}
	idleChan = make(chan int, *threshold)
	start := time.Now().UnixNano()
	var i int
	for q := 0; q < *threshold; q++ {
		idleChan <- 1
	}

	go func() {
		for range idleChan {
			config := SolverConfig{Prefix: *prefix, Difficulty: *difficulty}
			solver, _ := New(config)
			fmt.Println("[@]" + strconv.Itoa(i*k) + " Working...")
			go solver.Solve(i)
			i++
		}
	}()
	select {
	case result := <-resultChan:
		end := time.Now().UnixNano()
		fmt.Println("[+]Attenpts: ", result.attempts)
		fmt.Println("[*]Time spend: ", (int64((end-start)/int64(time.Nanosecond)) / int64(time.Second)))
		if *binString {
			fmt.Println("[+]binString: ", result.binString)
		}
	case <-time.After(5 * time.Minute):
		fmt.Println("timeout!")
		os.Exit(0)

	}
}
