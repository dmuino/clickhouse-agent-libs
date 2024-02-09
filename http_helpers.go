package chagent

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const HtmlHeader = `<!DOCTYPE html>
<html>
<header>
  <meta charset="utf-8">
</header>
<body bgcolor='#333333' style='font-weight: 500; font-size:14px; font-family: "Courier New", Courier, monospace'>
<pre>
`

func EtcHostsEndpoint(logger *Logger, w http.ResponseWriter, r *http.Request) {
	// return the contents of /etc/hosts
	contents, err := os.ReadFile("/etc/hosts")
	if err != nil {
		logger.Errorf("Error reading /etc/hosts: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	// write response
	w.Header().Set("Content-Type", "text/plain")
	n, err := w.Write(contents)
	if err != nil {
		logger.Warningf("Error writing /etc/hosts: %v", err)
	} else if n != len(contents) {
		logger.Warningf("Error writing /etc/hosts: partial write %d/%d", n, len(contents))
	} else {
		logger.Infof("Wrote /etc/hosts to client %s", r.RemoteAddr)
	}
}

func SendOutputFromCommand(logger *Logger, command []string, w http.ResponseWriter, r *http.Request) {
	// run a command and return the output
	cmd := exec.Command(command[0], command[1:]...)

	var errb, outb bytes.Buffer
	cmd.Stderr = &errb
	cmd.Stdout = &outb
	err := cmd.Run()
	if err != nil {
		logger.Errorf("Error running command %s: %v", cmd, err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"error":"%v"}`, err)))
		return
	}

	// write response
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(HtmlHeader))
	if err != nil {
		return
	}

	stdout := outb.String()
	stderr := errb.String()

	// send our standard html header

	// write stderr first using red color, then stdout using normal color
	_, _ = w.Write([]byte(fmt.Sprintf(`<span style="color:red">%s</span><span style="color:white">%s</span>`,
		stderr, stdout)))

	_, _ = w.Write([]byte("</pre></body></html>"))

	logger.Infof("Wrote output of %s to client %s", command, r.RemoteAddr)
}

func tailOutputFromCommandFmt(logger *Logger, command []string, w http.ResponseWriter, r *http.Request,
	formatter func(*Logger, string, string) string) {
	logger.Infof("Running command %v - will send output to %s", command, r.RemoteAddr)
	cmd := exec.Command(command[0], command[1:]...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logger.Errorf("Error getting stdout pipe for command %v: %v", cmd, err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"error":"%v"}`, err)))
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		logger.Errorf("Error getting stderr pipe for command %s: %v", cmd, err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"error":"%v"}`, err)))
		return
	}

	stdoutChan := make(chan string)
	stderrChan := make(chan string)

	toChannel := func(ch chan string, reader io.ReadCloser) {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			ch <- scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			logger.Errorf("Error reading stdout from command %v: %v", command, err)
		}
	}
	go toChannel(stdoutChan, stdout)
	go toChannel(stderrChan, stderr)

	err = cmd.Start()
	if err != nil {
		logger.Errorf("Error starting command %s: %v", cmd, err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"error":"%v"}`, err)))
		return
	}

	logger.Infof("Started command %s - will send output to %s", cmd, r.RemoteAddr)
	// write response
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(HtmlHeader))
	if err != nil {
		return
	}

	killAndRelease := func(cmd *exec.Cmd, client string, err error) {
		_ = cmd.Process.Kill()
		_ = cmd.Process.Release()
		logger.Infof("Wrote output of %s to client %s - %v", cmd, client, err)
	}

	// read from stdout and stderr channels and write to the response
	handleOutput := func(line string, color string) bool {
		htmlLine := formatter(logger, line, color)
		_, err = w.Write([]byte(htmlLine))
		if err != nil {
			killAndRelease(cmd, r.RemoteAddr, err)
			return true
		}
		// flush the response
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		return false
	}

	for {
		select {
		case line := <-stdoutChan:
			if handleOutput(line, "white") {
				return
			}
		case line := <-stderrChan:
			if handleOutput(line, "red") {
				return
			}
		case <-r.Context().Done():
			logger.Infof("Client %s disconnected - killing command %s", r.RemoteAddr, cmd)
			_ = cmd.Process.Kill()
			_ = cmd.Process.Release()
			return
		case <-time.After(10 * time.Second):
			// check if the command is still running
			if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
				logger.Infof("Command %s has exited", cmd)
				// clean up resources
				_ = cmd.Process.Release()
				_, err = w.Write([]byte("</pre></body></html>"))
				return
			} else {
				logger.Debugf("[%s] Command %s is still running", r.RemoteAddr, cmd)
			}
		}
	}
}

func getTextOutput(_ *Logger, line, color string) string {
	return fmt.Sprintf(`<span style="color:%s">%s%c</span>`, color, line, '\n')
}

func TailLogs(logger *Logger, logFile string, numLines int, w http.ResponseWriter, r *http.Request) {
	n := strconv.Itoa(numLines)
	tailOutputFromCommandFmt(logger, []string{"tail", "-F", "-n", n, logFile}, w, r, jsonLogFormatter)
}

func TailOutputFromCommand(logger *Logger, command []string, w http.ResponseWriter, r *http.Request) {
	tailOutputFromCommandFmt(logger, command, w, r, getTextOutput)
}

var levelsAsStr = []string{"FATAL", "CRITICAL", "ERROR", "WARN", "NOTICE", "INFO", "DEBUG", "TRACE"}
var colorsForLevel = []string{"red", "red", "red", "yellow", "white", "white", "green", "green"}

func toLevelColorStr(level string) (string, string) {
	num, err := strconv.Atoi(level)
	if err != nil {
		return level, "white"
	}
	if num < 1 || num > len(levelsAsStr) {
		return level, "white"
	}
	num--
	return levelsAsStr[num], colorsForLevel[num]
}

func jsonLogFormatter(logger *Logger, jsonLine string, color string) string {
	// parse jsonLine as map[string]interface{}
	var parsed map[string]interface{}
	err := json.Unmarshal([]byte(jsonLine), &parsed)
	if err != nil {
		logger.Warningf("Error parsing json line: %s - %v", jsonLine, err)
		return getTextOutput(logger, jsonLine, color)
	}

	// Example json line
	// {"date_time":"1701376786.499134","thread_name":"","thread_id":"4648","level":"6","query_id":"","logger_name":"DNSCacheUpdater","message":"IPs of some hosts have been changed. Will reload cluster config.","source_file":"src\/Interpreters\/DNSCacheUpdater.cpp; void DB::DNSCacheUpdater::run()","source_line":"27"}

	// parse date_time
	dateTimeStr := parsed["date_time"].(string)
	dateTimeSecs, err := strconv.ParseFloat(dateTimeStr, 64)
	if err != nil {
		logger.Warningf("Error parsing date_time: %s - %v", dateTimeStr, err)
		return getTextOutput(logger, jsonLine, color)
	}

	// convert fractional seconds (includes microseconds) to time.Time
	dateTime := time.Unix(int64(dateTimeSecs), int64((dateTimeSecs-float64(int64(dateTimeSecs)))*1e9))

	threadId := parsed["thread_id"].(string)
	level, color := toLevelColorStr(parsed["level"].(string))

	msg := strings.TrimSpace(parsed["message"].(string))
	lineEnd := "</span>\n"

	return fmt.Sprintf(`<span style="color:%s">%s %s [%s] [%s] %s <%s:%s> [%s]%s`,
		color,
		dateTime.Format("2006-01-02T15:04:05.000"),
		level, threadId,
		parsed["logger_name"],
		msg,
		parsed["source_file"],
		parsed["source_line"],
		parsed["query_id"],
		lineEnd)
}
