package ipfs

import (
	"context"
	"fmt"
	"io"

	shell "github.com/ipfs/go-ipfs-api"
	myio "gitlink.org.cn/cloudream/common/utils/io"
)

type Client struct {
	shell *shell.Shell
}

func NewClient(cfg *Config) (*Client, error) {
	ipfsAddr := fmt.Sprintf("localhost:%d", cfg.Port)
	sh := shell.NewShell(ipfsAddr)

	// 检测连通性
	if !sh.IsUp() {
		return nil, fmt.Errorf("cannot connect to %s", ipfsAddr)
	}

	return &Client{
		shell: sh,
	}, nil
}

func (fs *Client) IsUp() bool {
	return fs.shell.IsUp()
}

func (fs *Client) CreateFileStream() (myio.PromiseWriteCloser[string], error) {
	pr, pw := io.Pipe()

	ipfsWriter := ipfsWriter{
		writer:   pw,
		finished: make(chan any, 1),
	}

	go func() {
		hash, err := fs.shell.Add(pr)
		ipfsWriter.finishErr = err
		ipfsWriter.fileHash = hash
		close(ipfsWriter.finished)
		pr.CloseWithError(err)
	}()

	return &ipfsWriter, nil
}

func (fs *Client) CreateFile(file io.Reader) (string, error) {
	return fs.shell.Add(file)
}

func (fs *Client) OpenRead(hash string) (io.ReadCloser, error) {
	return fs.shell.Cat(hash)
}

func (fs *Client) Pin(hash string) error {
	return fs.shell.Pin(hash)
}

func (fs *Client) Unpin(hash string) error {
	return fs.shell.Unpin(hash)
}

func (fs *Client) GetPinnedFiles() (map[string]shell.PinInfo, error) {
	return fs.shell.PinsOfType(context.Background(), shell.RecursivePin)
}

func (fs *Client) List(hash string) ([]*shell.LsLink, error) {
	return fs.shell.List(hash)
}

type ipfsWriter struct {
	writer    *io.PipeWriter
	finished  chan any
	finishErr error
	fileHash  string
}

func (p *ipfsWriter) Write(data []byte) (n int, err error) {
	return p.writer.Write(data)
}

// 设置一个error中断写入
func (w *ipfsWriter) Abort(err error) {
	w.writer.CloseWithError(err)
}

// Finish 结束写入，并获得返回值（文件哈希值）
func (w *ipfsWriter) Finish() (string, error) {
	w.writer.CloseWithError(io.EOF)

	<-w.finished

	return w.fileHash, w.finishErr
}