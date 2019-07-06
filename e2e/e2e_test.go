package e2e

import (
	"bufio"
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/DATA-DOG/godog"
	"github.com/docker/docker/api/types"
	reg "github.com/genuinetools/reg/registry"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/database/mysql"
	_ "github.com/golang-migrate/migrate/source/file"
	digest "github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"
)

var (
	httpResponse      *http.Response
	imagespyAPICmd    *exec.Cmd
	imagespyAPICmdOut *bytes.Buffer
)

type testingT struct {
	args   []interface{}
	format string
}

func (t *testingT) Errorf(format string, args ...interface{}) {
	t.args = args
	t.format = format
}

func (t *testingT) getLastError() error {
	return fmt.Errorf(t.format, t.args...)
}

type timeMock struct {
	callCount int
}

func (m *timeMock) Time() time.Time {
	m.callCount++
	t, _ := time.Parse("2006-01-02 15:04", fmt.Sprintf("2018-10-26 %02d:00", m.callCount))
	return t.UTC()
}

func aCleanDatabase() error {
	connection := "root:root@tcp(127.0.0.1:33306)/imagespy?charset=utf8&parseTime=True&loc=UTC"
	db, err := sql.Open("mysql", connection)
	if err != nil {
		return fmt.Errorf("Unable to connect to database: %s", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return fmt.Errorf("Unable to ping database before test: %s", err)
	}

	_, err = db.Exec("DROP DATABASE IF EXISTS imagespy")
	if err != nil {
		return fmt.Errorf("Unable to drop database before test: %s", err)
	}

	_, err = db.Exec("CREATE DATABASE imagespy")
	if err != nil {
		return fmt.Errorf("Unable to create database before test: %s", err)
	}

	return nil
}

func theDockerImagePushed(name string) error {
	pushCmd := exec.Command("docker", "push", name)
	pushOut, err := pushCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Unable to push Docker image %s: %s", name, string(pushOut))
	}

	return nil
}

func sendingARequestTo(method, path string) error {
	req, err := http.NewRequest(method, "http://127.0.0.1:3001"+path, nil)
	if err != nil {
		return fmt.Errorf("creating request: %s", err)
	}

	httpResponse, err = http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending request to %s: %s", path, err)
	}

	return nil
}

func imagespyAPIStarted() error {
	imagespyAPICmd = exec.Command(
		"../api",
		"server",
		"--db.connection", "root:root@tcp(127.0.0.1:33306)/imagespy?charset=utf8&parseTime=True&loc=UTC",
		"--http.address", "127.0.0.1:3001",
		"--log.level", "debug",
		"--migrations.enabled",
		"--migrations.path", "file://../store/gorm/migrations",
		"--registry.address", "127.0.0.1:52854",
		"--registry.insecure",
	)

	imagespyAPICmdOut = bytes.NewBufferString("")
	imagespyAPICmd.Stdout = imagespyAPICmdOut
	imagespyAPICmd.Stderr = imagespyAPICmdOut
	err := imagespyAPICmd.Start()
	if err != nil {
		return fmt.Errorf("starting imagespy API: %s", err)
	}

	for i := 0; i < 20; i++ {
		time.Sleep(100 * time.Millisecond)
		conn, _ := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", "3001"), 10*time.Millisecond)
		if conn != nil {
			conn.Close()
			return nil
		}
	}

	return fmt.Errorf("unable to connect to imagespy API after 20 tries")
}

func aCleanRegistry() error {
	r, err := reg.New(
		types.AuthConfig{ServerAddress: "127.0.0.1:52854"},
		reg.Opt{Insecure: true, SkipPing: true},
	)
	if err != nil {
		return err
	}

	repositories, err := r.Catalog("")
	if err != nil {
		return err
	}

	for _, repo := range repositories {
		digestDedup := make(map[string]digest.Digest)
		tags, err := r.Tags(repo)
		if err != nil {
			return err
		}

		for _, t := range tags {
			d, err := r.Digest(reg.Image{
				Path: repo,
				Tag:  t,
			})
			if err != nil {
				return err
			}

			_, ok := digestDedup[d.String()]
			if !ok {
				digestDedup[d.String()] = d
			}
		}

		for _, d := range digestDedup {
			err := r.Delete(repo, d)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func theDockerImageTaggedAs(sourceRef, targetRef string) error {
	pushCmd := exec.Command("docker", "tag", sourceRef, targetRef)
	pushOut, err := pushCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Unable to tag Docker image %s as %s: %s", sourceRef, targetRef, string(pushOut))
	}

	return nil
}

func theAPIRespondsWithStatusCode(code string) error {
	statusCode, err := strconv.Atoi(code)
	if err != nil {
		return fmt.Errorf("parsing status code %s: %s", code, err)
	}

	if statusCode != httpResponse.StatusCode {
		return fmt.Errorf("expected status code %d got %d", statusCode, httpResponse.StatusCode)
	}

	return nil
}

func theAPIRespondsWithABodyOf(fixtureFilePath string) error {
	t := &testingT{}
	f, err := os.Open(fixtureFilePath)
	if err != nil {
		return fmt.Errorf("opening fixture file: %s", err)
	}

	defer f.Close()
	expectedBytes, err := ioutil.ReadAll(f)
	if err != nil {
		return fmt.Errorf("reading fixture file: %s", err)
	}

	defer httpResponse.Body.Close()
	actualBytes, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %s", err)
	}

	if !assert.JSONEq(t, string(expectedBytes), string(actualBytes)) {
		return t.getLastError()
	}

	return nil
}

func sendingTheRequestTo(requestFilePath, path string) error {
	f, err := os.Open(requestFilePath)
	if err != nil {
		return fmt.Errorf("opening request fixture file: %s", err)
	}

	defer f.Close()
	req, err := http.ReadRequest(bufio.NewReader(f))
	if err != nil {
		return fmt.Errorf("creating request from file: %s", err)
	}

	u, err := url.Parse("http://127.0.0.1:3001" + path)
	if err != nil {
		return fmt.Errorf("parsing request url: %s", err)
	}

	req.RequestURI = ""
	req.URL = u
	httpResponse, err = http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending request from file to %s: %s", path, err)
	}

	return nil
}

func waitingFor(s string) error {
	d, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("parsing duration: %s", err)
	}

	time.Sleep(d)
	return nil
}

func runningTheUpdaterComand(cmdName string) error {
	cmd := exec.Command(
		"../api",
		"updater",
		cmdName,
		"--db.connection", "root:root@tcp(127.0.0.1:33306)/imagespy?charset=utf8&parseTime=True&loc=UTC",
		"--log.level", "debug",
		"--registry.address", "127.0.0.1:52854",
		"--registry.insecure",
	)
	stdErrOut, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("running the updater: %s\n\n%s", err, string(stdErrOut))
	}

	return nil
}

func FeatureContext(s *godog.Suite) {
	s.Step(`^the Docker image "([^"]*)" pushed$`, theDockerImagePushed)
	s.Step(`^sending a "([^"]*)" request to "([^"]*)"$`, sendingARequestTo)
	s.Step(`^imagespy API started$`, imagespyAPIStarted)
	s.Step(`^the Docker image "([^"]*)" tagged as "([^"]*)"$`, theDockerImageTaggedAs)
	s.Step(`^the API responds with status code "([^"]*)"$`, theAPIRespondsWithStatusCode)
	s.Step(`^the API responds with a body of "([^"]*)"$`, theAPIRespondsWithABodyOf)
	s.Step(`^running the updater comand "([^"]*)"$`, runningTheUpdaterComand)
	s.Step(`^sending the request "([^"]*)" to "([^"]*)"$`, sendingTheRequestTo)
	s.Step(`^waiting for "([^"]*)"$`, waitingFor)

	s.BeforeScenario(func(interface{}) {
		err := aCleanDatabase()
		if err != nil {
			panic(err)
		}

		err = aCleanRegistry()
		if err != nil {
			panic(err)
		}
	})

	s.AfterScenario(func(_ interface{}, scErr error) {
		if imagespyAPICmd != nil {
			if scErr != nil {
				fmt.Println(imagespyAPICmdOut)
			}

			err := imagespyAPICmd.Process.Kill()
			if err != nil {
				panic(err)
			}
		}
	})
}
