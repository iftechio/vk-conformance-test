package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"k8s.io/client-go/tools/clientcmd"

	"github.com/iftechio/vk-test/suite"
	"github.com/iftechio/vk-test/testcases"
)

type testResult struct {
	Name     string
	Desc     string
	Duration time.Duration
	Err      error
}

func main() {
	var (
		nodeName   string
		kubeconfig string
		timeout    time.Duration
		runTests   string
	)

	availTests := make([]string, 0, len(testcases.AvailableTests))
	for name := range testcases.AvailableTests {
		availTests = append(availTests, name)
	}

	flag.StringVar(&kubeconfig, "kubeconfig", filepath.Join(os.Getenv("HOME"), ".kube", "config"), "Path to the kubeconfig file to use for CLI requests.")
	flag.StringVar(&runTests, "run", "", fmt.Sprintf("Regexp to match test names. Available tests: %s", strings.Join(availTests, ",")))
	flag.StringVar(&nodeName, "nodename", "virtual-kubelet", "Target node name")
	flag.DurationVar(&timeout, "test.timeout", time.Minute*5, "Timeout of each test")
	flag.Parse()

	restCfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalln(err)
	}
	s, err := suite.New(restCfg, nodeName)
	if err != nil {
		log.Fatalf("create test suite: %s", err)
	}

	var (
		wg            sync.WaitGroup
		expectedTests []testcases.Tester
		pattern       *regexp.Regexp
		rootCtx       = context.TODO()
	)
	if len(runTests) != 0 {
		pattern, err = regexp.Compile(runTests)
		if err != nil {
			log.Fatalf("invalid regexp: %s: %s", runTests, err)
		}
	} else {
		pattern = regexp.MustCompile(".*")
	}
	for name, t := range testcases.AvailableTests {
		if pattern.MatchString(name) {
			expectedTests = append(expectedTests, t)
		}
	}
	n := len(expectedTests)
	wg.Add(n)
	resultCh := make(chan testResult, n)
	fmt.Printf("Running %d tests...\n", n)
	for _, t := range expectedTests {
		go func(t testcases.Tester) {
			defer wg.Done()
			ctx, cancelF := context.WithTimeout(rootCtx, timeout)
			defer cancelF()
			begin := time.Now()
			err := t.Test(ctx, s)
			resultCh <- testResult{
				Name:     t.Name(),
				Desc:     t.Description(),
				Duration: time.Since(begin),
				Err:      err,
			}
		}(t)
	}
	wg.Wait()
	close(resultCh)

	var (
		succList, failList []testResult
	)
	for r := range resultCh {
		if r.Err == nil {
			succList = append(succList, r)
		} else {
			failList = append(failList, r)
		}
	}
	sortByName := func(results []testResult) {
		sort.Slice(results, func(i, j int) bool {
			return results[i].Name < results[j].Name
		})
	}
	sortByName(succList)
	sortByName(failList)

	if len(succList) > 0 {
		fmt.Printf("====== %d Test(s) passed\n", len(succList))
		for _, r := range succList {
			fmt.Printf("-----\t%s (%s)\n", r.Name, r.Duration)
		}
	}
	if len(failList) > 0 {
		fmt.Printf("====== %d Test(s) failed\n", len(failList))
		for _, r := range failList {
			fmt.Printf("-----\t%s (%s)\n", r.Name, r.Duration)
			fmt.Printf(">>>>>\t%s\n", r.Desc)
			fmt.Println(r.Err.Error())
		}
	}
}
