package main

import (
	"errors"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
)

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////
const (
	porterrmsg = "Invalid port specification"
)

func dashSplit(sp string, ports *[]int) error {
	dp := strings.Split(sp, "-")
	if len(dp) != 2 {
		return errors.New(porterrmsg)
	}
	start, err := strconv.Atoi(dp[0])
	if err != nil {
		return errors.New(porterrmsg)
	}
	end, err := strconv.Atoi(dp[1])
	if err != nil {
		return errors.New(porterrmsg)
	}
	if start > end || start < 1 || end > 65535 {
		return errors.New(porterrmsg)
	}
	for ; start <= end; start++ {
		*ports = append(*ports, start)
	}
	return nil
}

func convertAndAddPort(p string, ports *[]int) error {
	i, err := strconv.Atoi(p)
	if err != nil {
		return errors.New(porterrmsg)
	}
	if i < 1 || i > 65535 {
		return errors.New(porterrmsg)
	}
	*ports = append(*ports, i)
	return nil
}

// Parse turns a string of ports separated by '-' or ',' and returns a slice of Ints.
// конвертирует готовый стринг портов в интовый срез
func Parse(s string) ([]int, error) {
	ports := []int{}
	if strings.Contains(s, ",") && strings.Contains(s, "-") {
		sp := strings.Split(s, ",")
		for _, p := range sp {
			if strings.Contains(p, "-") {
				if err := dashSplit(p, &ports); err != nil {
					return ports, err
				}
			} else {
				if err := convertAndAddPort(p, &ports); err != nil {
					return ports, err
				}
			}
		}
	} else if strings.Contains(s, ",") {
		sp := strings.Split(s, ",")
		for _, p := range sp {
			convertAndAddPort(p, &ports)
		}
	} else if strings.Contains(s, "-") {
		if err := dashSplit(s, &ports); err != nil {
			return ports, err
		}
	} else {
		if err := convertAndAddPort(s, &ports); err != nil {
			return ports, err
		}
	}
	return ports, nil
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func worker(ports chan int, results chan int) {
	for p := range ports {
		adress := fmt.Sprintf("scanme.nmap.org:%d", p)
		conn, err := net.Dial("tcp", adress)
		if err != nil {
			results <- 0
			continue
		}
		conn.Close()
		results <- p
	}

}

func main() {
	ports := make(chan int, 250)
	results := make(chan int)
	var openports []int
	for i := 0; i < cap(ports); i++ {
		go worker(ports, results)
		//последовательно для канала результата сохраняет открытые порты
	}

	//заполнение канала портс для работы функции worker -> ждет
	go func() {
		for i := 1; i <= 1024; i++ {
			ports <- i
		}
	}()

	for i := 0; i < 1024; i++ {
		port := <-results
		//последовательная передача херовенек из канала резалтс
		if port != 0 {
			openports = append(openports, port)
		}
	}
	close(ports)
	close(results)
	sort.Ints(openports)
	for _, port := range openports {
		fmt.Printf("%d open\n", port)
	}
}
