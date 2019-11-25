package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"k8s.io/client-go/tools/clientcmd"

	"github.com/iftechio/vk-test/suite"
	"github.com/iftechio/vk-test/testcases"
)

type testResult struct {
	Name     string
	Duration time.Duration
	Err      error
}

func main() {
	var (
		nodeName   string
		kubeconfig string
		timeout    time.Duration
	)

	flag.StringVar(&kubeconfig, "kubeconfig", filepath.Join(os.Getenv("HOME"), ".kube", "config"), "Path to the kubeconfig file to use for CLI requests.")
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
	var wg sync.WaitGroup
	rootCtx := context.TODO()
	n := len(testcases.AvailableTests)
	wg.Add(n)
	resultCh := make(chan testResult, n)
	fmt.Printf("Running %d tests...\n", n)
	for name, t := range testcases.AvailableTests {
		go func(name string, t testcases.Tester) {
			defer wg.Done()
			ctx, cancelF := context.WithTimeout(rootCtx, timeout)
			defer cancelF()
			begin := time.Now()
			err := t.Test(ctx, s)
			resultCh <- testResult{
				Name:     name,
				Duration: time.Since(begin),
				Err:      err,
			}
		}(name, t)
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
			fmt.Println(r.Err.Error())
		}
	}
}
