package sshkit

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/oklog/ulid"
	"github.com/zhiminwen/magetool/fmtkit"
	"golang.org/x/crypto/ssh"
)

var myfmt fmtkit.Formatter

func init() {
	myfmt = &fmtkit.BasicFormatter{}
}

func SetFormatter(fmt fmtkit.Formatter) {
	myfmt = fmt
}

type SSHClient struct {
	Host           string
	Port           string
	User           string
	Password       string
	PrivatekeyFile string
	UUID           string

	Properties map[string]string

	ClientConfig *ssh.ClientConfig
	// DialTimeoutSecond int
	// MaxDataThroughput uint64

	sshClient *ssh.Client
	// sshSession  *ssh.Session
	isConnected bool
}

func (c *SSHClient) GetSSHClient() *ssh.Client {
	return c.sshClient
}

func AuthByPrivateKey(keyfile string) (ssh.AuthMethod, error) {
	pemBytes, err := os.ReadFile(keyfile)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(pemBytes)
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(signer), nil
}

func NewSSHClient(host, port, user, password, keyfile string) (*SSHClient, error) {
	if password == "" && keyfile == "" {
		log.Printf("Failed to construct ssh client. both password and private key are empty.")
		return nil, fmt.Errorf("failed to construct ssh client. both password and private key are empty")
	}

	var authMethod ssh.AuthMethod
	var err error

	if password != "" {
		authMethod = ssh.Password(password)
	}
	if keyfile != "" {
		authMethod, err = AuthByPrivateKey(keyfile)
		if err != nil {
			log.Printf("Failed to get public keys from supplied keyfile. Error: %v", err)
			return nil, err
		}
	}
	now := time.Now()
	source := now.UnixNano() + rand.Int63()
	entropy := ulid.Monotonic(rand.New(rand.NewSource(source)), 0)
	uuid := ulid.MustNew(ulid.Timestamp(now), entropy)

	client := &SSHClient{
		Host:           host,
		Port:           port,
		User:           user,
		Password:       password,
		Properties:     map[string]string{},
		PrivatekeyFile: keyfile,
		ClientConfig: &ssh.ClientConfig{
			User: user,
			Auth: []ssh.AuthMethod{
				authMethod,
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		},

		UUID: uuid.String(),
	}

	return client, nil
}

func (c *SSHClient) SetProperty(key, value string) {
	c.Properties[key] = value
}

func (c *SSHClient) Connect() error {
	if c.isConnected {
		return nil
	}

	sshClient, err := ssh.Dial("tcp", c.Host+":"+c.Port, c.ClientConfig)

	if err != nil {
		log.Printf("Failed to connect to %s. error:%v", c.Host, err)
		return err
	}

	c.sshClient = sshClient
	c.isConnected = true

	return nil
}

func (c *SSHClient) NewSession() (*ssh.Session, error) {
	err := c.Connect()
	if err != nil {
		return nil, err
	}

	session, err := c.sshClient.NewSession()
	if err != nil {
		log.Printf("Failed to create SSH session for %s. error:%v", c.Host, err)
		return nil, err
	}

	return session, nil
}

func (c *SSHClient) Close() error {
	if !c.isConnected {
		return nil
	}

	err := c.sshClient.Close()
	if err != nil {
		log.Printf("Failed to close SSH connection")
		return err
	}

	return nil
}

func (c *SSHClient) Capture(cmd string) (string, error) {
	session, err := c.NewSession()
	if err != nil {
		log.Printf("Failed to create session:%v", err)
		return "", err
	}
	defer session.Close()

	out, err := session.CombinedOutput(cmd)

	if err != nil {
		log.Printf("Failed to execute:%v", err)
		return "", err
	}

	result := strings.TrimSpace(string(out[:]))
	return result, nil
}

func (c *SSHClient) display(reader *bufio.Reader, stderr bool, wg *sync.WaitGroup) {
	prefix := fmt.Sprintf("%s (%s:%s) :", c.UUID, c.Host, c.Port)
	//Bufio.Scanner has a read buf limit: 64k. revert back to readline
	printLine := func(prefix string, line []byte) {
		if stderr {
			myfmt.ErrorLine(prefix, string(line[:]))
		} else {
			myfmt.NormalLine(prefix, string(line[:]))
		}
	}

	var err error
	var line []byte
	for {
		line, _, err = reader.ReadLine()
		if err != nil {
			break
		}
		// ReadLine either returns a non-nil line or it returns an error, never both.
		printLine(prefix, line)
	}

	if err != io.EOF {
		myfmt.ErrorLine(prefix, fmt.Sprintf("error: %v", err))
	}
	wg.Done()
}

func (c *SSHClient) MustExecute(cmd string) {
	err := c.Execute(cmd)
	if err != nil {
		log.Fatalf("Execution fatal failure")
	}
}

func (c *SSHClient) Execute(cmd string) error {
	var wg sync.WaitGroup

	myfmt.Header(cmd)
	startTime := time.Now()

	session, err := c.NewSession()
	if err != nil {
		log.Printf("Failed to create session:%v", err)
		return err
	}
	defer session.Close()

	//pipe need to be before Start
	outReader, err := session.StdoutPipe()
	if err != nil {
		log.Printf("error on getting stdout pipe:%v", err)
		myfmt.Footer(time.Since(startTime), err)
		return err
	}
	outLineReader := bufio.NewReader(outReader)

	errReader, err := session.StderrPipe()
	if err != nil {
		log.Printf("error on getting stderr pipe:%v", err)
		myfmt.Footer(time.Since(startTime), err)
		return err
	}
	errLineReader := bufio.NewReader(errReader)

	wg.Add(1)
	go c.display(outLineReader, false, &wg)

	wg.Add(1)
	go c.display(errLineReader, true, &wg)

	err = session.Start(cmd)
	if err != nil {
		log.Printf("error on starting the session:%v", err)
		myfmt.Footer(time.Since(startTime), err)
		return err
	}

	err = session.Wait()
	if err != nil {
		// log.Printf("error on session wait:%v", err)
		myfmt.Footer(time.Since(startTime), err)
		return err
	}

	wg.Wait()
	myfmt.Footer(time.Since(startTime), err)
	return nil
}

func (c *SSHClient) RequestPty(session *ssh.Session) error {
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}
	// Request pseudo terminal
	if err := session.RequestPty("xterm", 40, 80, modes); err != nil {
		log.Printf("request for pseudo terminal failed: %v", err)
		return err
	}
	return nil
}

func (c *SSHClient) ExecuteInteractively(cmd string, inputMap map[string]string) error {
	myfmt.Header(cmd)
	startTime := time.Now()

	session, err := c.NewSession()
	if err != nil {
		log.Printf("Failed to create session:%v", err)
		return err
	}
	defer session.Close()

	err = c.RequestPty(session)
	if err != nil {
		log.Printf("Failed to request pty")
		return err
	}

	//pipe need to be before Start
	stdin, err := session.StdinPipe()
	if err != nil {
		log.Printf("error on getting stdin pipe:%v", err)
		myfmt.Footer(time.Since(startTime), err)
		return err
	}

	reader, err := session.StdoutPipe()
	if err != nil {
		log.Printf("error on getting stdout pipe:%v", err)
		myfmt.Footer(time.Since(startTime), err)
		return err
	}
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanBytes)

	err = session.Start(cmd)
	if err != nil {
		log.Printf("error on starting the session:%v", err)
		myfmt.Footer(time.Since(startTime), err)
		return err
	}

	go func() {
		var line strings.Builder
		prefix := fmt.Sprintf("%s (%s:%s) :", c.UUID, c.Host, c.Port)
		for scanner.Scan() {
			b := scanner.Text()
			if b == "\n" {
				myfmt.NormalLine(prefix, strings.TrimRight(line.String(), "\r"))
				line.Reset()
			}
			line.WriteString(b)
			for pattern, text := range inputMap {
				reg := regexp.MustCompile(pattern)
				if reg.MatchString(line.String()) {
					fmt.Fprintln(stdin, text)
				}
			}
		}
		err := scanner.Err()
		if err != nil {
			myfmt.ErrorLine(prefix, fmt.Sprintf("error: %v", err))
		}
	}()

	err = session.Wait()
	if err != nil {
		// log.Printf("error on session wait:%v", err)
		myfmt.Footer(time.Since(startTime), err)
		return err
	}

	myfmt.Footer(time.Since(startTime), err)
	return nil
}

func (c *SSHClient) UploadByReader(r io.Reader, remoteFullPath string, size int64, permission string) error {
	session, err := c.NewSession()
	if err != nil {
		log.Printf("Failed to create session:%v", err)
		return err
	}
	defer session.Close()
	w, err := session.StdinPipe()
	if err != nil {
		log.Printf("Failed to create stdin pipe: %v", err)
		return err
	}

	defer w.Close()

	// session.Stdout = os.Stdout //for debugging only

	err = session.Start("/usr/bin/scp -qt " + path.Dir(remoteFullPath))
	if err != nil {
		log.Printf("Failed to start session:%v", err)
		return err
	}

	go func() {
		iop := NewIOProgress(size, "Uploading", "Uploaded")
		// teeReader := io.TeeReader(r, iop.Bar)

		fmt.Fprintln(w, "C"+permission, size, path.Base(remoteFullPath))
		// bytes, err := io.Copy(w, teeReader)
		// bytes, err := io.Copy(io.MultiWriter(w, iop.Bar), r)
		_, err := io.Copy(io.MultiWriter(w, iop.Bar), r)
		// _, err := io.Copy(w, r)
		if err != nil {
			log.Printf("Failed to copy io: %v", err)
		}
		fmt.Fprintln(w, "\x00")
		// log.Printf("its done. copied %d bytes", bytes)
		iop.Bar.Close() //exit to stop the updating of the bar
		// iop.FinalMessage()
	}()

	err = session.Wait()
	if err != nil {
		if err.Error() == "Process exited with status 1" {
			return nil //consider as success?
		}
		log.Printf("error on session wait:%v", err)
		return err
	}

	return nil
}

func (c *SSHClient) Upload(localFullPath, remoteFullPath string) error {
	file, err := os.Open(localFullPath)
	if err != nil {
		log.Printf("Failed to open local file:%v", err)
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		log.Printf("Failed to stat the local file:%v", err)
		return err
	}
	// use bufio incase the file is big
	r := bufio.NewReader(file)
	return c.UploadByReader(r, remoteFullPath, stat.Size(), "0660")
}

func (c *SSHClient) Put(content string, remoteFullPath string) error {
	r := strings.NewReader(content)
	return c.UploadByReader(r, remoteFullPath, int64(len(content)), "0600")
}

func (c *SSHClient) DownloadByWriter(remoteFullPath string, dstWriter io.Writer) error {
	session, err := c.NewSession()
	if err != nil {
		log.Printf("Failed to create session:%v", err)
		return err
	}
	defer session.Close()

	w, err := session.StdinPipe()
	if err != nil {
		log.Printf("Failed to create stdin pipe: %v", err)
		return err
	}
	defer w.Close()

	r, err := session.StdoutPipe()
	if err != nil {
		log.Printf("Failed to create stdout pipe: %v", err)
		return err
	}

	//"-f"is sink mode
	err = session.Start("/usr/bin/scp -f " + remoteFullPath)
	if err != nil {
		log.Printf("Failed to start session:%v", err)
		return err
	}

	go sinkProtocol(r, w, dstWriter)

	err = session.Wait()
	if err != nil {
		if err.Error() == "Process exited with status 1" {
			return nil //consider as success?
		}
		log.Printf("error on session wait:%v", err)
		return err
	}

	return nil
}

func (c *SSHClient) Download(remoteFullPath, localFullPath string) error {
	f, err := os.Create(localFullPath)
	if err != nil {
		log.Printf("Failed to create file")
		return err
	}

	defer f.Close()

	w := bufio.NewWriter(f)

	err = c.DownloadByWriter(remoteFullPath, w)
	if err != nil {
		return err
	}
	w.Flush()

	return nil
}

func (c *SSHClient) Get(remoteFullPath string) (string, error) {
	b := &strings.Builder{}

	err := c.DownloadByWriter(remoteFullPath, b)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}
