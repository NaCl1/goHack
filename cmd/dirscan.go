package cmd

import (
	"bufio"
	"fmt"
	"github.com/imroc/req/v3"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	urls     []string
	wordlist []string
	redirect bool
	timeout  int64
	threads  int
	cookie   string
	output   string
	delay    int
)

type outputType struct {
	path       string
	statusCode int
	redirect   string
	resLength  int64
}

func scanThreadWorker(client *req.Client, targetUrl string, path string, outputCh chan outputType, bar *progressbar.ProgressBar) {
	log.Debug(targetUrl, path, " scanThreadWorker")
	var oT outputType
	oT.path = path
	res, err := client.R().Get(targetUrl + path)
	if err != nil {
		log.Warn("[*] ", path, " error: ", err.Error())
		return
	}

	oT.resLength = res.ContentLength
	oT.statusCode = res.StatusCode
	oT.redirect = res.Header.Get("Location")

	if oT.statusCode/100 == 2 {
		outputCh <- oT
	}
	if oT.statusCode/100 == 3 {
		outputCh <- oT
	}
	if oT.statusCode/100 == 4 {
		outputCh <- oT
	}
	if oT.statusCode/100 == 5 {
		outputCh <- oT
	}
	printWorker(oT, bar)

	//fmt.Fprint(os.Stdout, "\r\033[K")
	//// 打印结果
	//fmt.Println(oT.path)
	//// 重新渲染进度条
	//bar.RenderBlank()
	//_ = bar.Add(1)

}

var result []outputType

func outputWorker(output outputType, bar *progressbar.ProgressBar) {
	result = append(result, output)
	//outputPrintWorker(output, bar)
}

func printWorker(output outputType, bar *progressbar.ProgressBar) {

	fmt.Fprint(os.Stdout, "\r\033[K")
	// 打印结果
	if output.statusCode != 404 {
		if output.redirect != "" {
			fmt.Println(output.statusCode, " - ", output.resLength, " - ", output.path, " -> ", output.redirect)
		} else {
			fmt.Println(output.statusCode, " - ", output.resLength, " - ", output.path)
		}

	}

	// 重新渲染进度条
	bar.RenderBlank()
	_ = bar.Add(1)
}

func countLines(filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		count++
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	return count, nil
}
func dirscan() interface{} {
	//defer close(jobs)
	//defer close(outputCh)
	// 禁止跟随跳转
	client := req.C().
		SetRedirectPolicy(func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}).
		SetTimeout(time.Duration(timeout) * time.Second) // 设置请求超时时间
	currentDir, err := os.Getwd()
	if err != nil {
		log.Error(err)
		return nil
	}
	for _, url := range urls {
		jobs := make(chan string, threads*2)
		outputCh := make(chan outputType, threads*2)

		log.Info("[*] dirscaning: ", url)
		if len(wordlist) > 0 {
			log.Info("[*] using wordlist: ", wordlist)
		} else {
			wordlist = append(wordlist, filepath.Clean(currentDir+"/"+"src/wordlists/top7000.txt"))
			log.Info("[*] using default wordlist: ", wordlist)
		}

		var outputWg sync.WaitGroup

		// 统计字典条数
		var lineCnt int
		for _, wordPath := range wordlist {
			cnt, _ := countLines(wordPath)
			lineCnt += cnt
		}

		bar := progressbar.NewOptions(lineCnt,
			progressbar.OptionSetDescription("Scanning"),
			progressbar.OptionShowCount(),
			progressbar.OptionSetWidth(30),
			progressbar.OptionSetPredictTime(true),
			progressbar.OptionSetTheme(progressbar.Theme{
				Saucer:        "=",
				SaucerPadding: "-",
				BarStart:      "[",
				SaucerHead:    ">",
				BarEnd:        "]",
			}),
		)

		// 启动输出线程
		outputWg.Add(1)
		go func() {
			defer outputWg.Done()
			for o := range outputCh {
				outputWorker(o, bar)
			}
		}()

		// 添加扫描线程
		var scanWg sync.WaitGroup
		for i := 0; i < threads; i++ {
			log.Debug("[*] scan thread ", i, " is added!")
			scanWg.Add(1)
			go func(id int) {
				defer scanWg.Done()
				for path := range jobs {
					scanThreadWorker(client, url, path, outputCh, bar)
				}
			}(i)

			// TODO

		}
		//scanWg.Add(threads)
		//for i := 0; i < threads; i++ {
		//
		//}
		// 逐条读取字典
		for _, wordlst := range wordlist {
			cnt := 0
			//file, err := os.OpenFile(wordlst, os.O_RDONLY, 0644)
			file, err := os.Open(wordlst)
			defer file.Close()
			if err != nil {
				log.Error(err)
			}
			scanner := bufio.NewScanner(file)

			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line != "" {

					jobs <- line
					cnt++
					//if cnt >= threads {
					//	//scanWg.Wait()
					//	time.Sleep(time.Duration(delay) * time.Millisecond)
					//}
				}
			}
			if err := scanner.Err(); err != nil {
				log.Error(err)
			}

		}

		// 任务发完后关闭 jobs
		close(jobs)

		// 等待所有扫描线程结束
		scanWg.Wait()

		// 然后关闭 outputCh，通知输出线程退出
		close(outputCh)

		// 等待输出线程结束
		outputWg.Wait()
		fmt.Print("\n")

		log.Info("[*] 扫描完成: ", url)

	}
	return nil
}

var dirscanCmd = &cobra.Command{
	Use:   "dirscan",
	Short: "目录扫描",
	Run: func(cmd *cobra.Command, args []string) {
		if len(urls) == 0 {
			cmd.Help()
			return
		}
		dirscan()
	},
}

func init() {
	log.SetFormatter(&log.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})
	log.SetLevel(log.InfoLevel)
	dirscanCmd.Flags().StringSliceVarP(&urls, "urls", "u", nil, "目标url")
	dirscanCmd.Flags().StringSliceVarP(&wordlist, "wordlist", "w", nil, "指定字典文件")
	dirscanCmd.Flags().Int64VarP(&timeout, "timeout", "", 5.0, "指定超时时间")
	dirscanCmd.Flags().BoolVarP(&redirect, "redirect", "", false, "设置是否跟随重定向")
	dirscanCmd.Flags().IntVarP(&threads, "threads", "t", 20, "指定线程数")
	dirscanCmd.Flags().StringVarP(&cookie, "cookie", "c", "", "指定cookie")
	dirscanCmd.Flags().StringVarP(&output, "output", "o", "", "指定输出文件路径")
	dirscanCmd.Flags().IntVarP(&delay, "delay", "d", 0, "延时(单位: 毫秒)")
	rootCmd.AddCommand(dirscanCmd)
}
