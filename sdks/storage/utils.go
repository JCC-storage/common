package cdssdk

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	myhttp "gitlink.org.cn/cloudream/common/utils/http"
	"gitlink.org.cn/cloudream/common/utils/math2"
	"gitlink.org.cn/cloudream/common/utils/serder"
)

func MakeIPFSFilePath(fileHash string) string {
	return filepath.Join("ipfs", fileHash)
}

func ParseJSONResponse[TBody any](resp *http.Response) (TBody, error) {
	var ret TBody
	contType := resp.Header.Get("Content-Type")
	if strings.Contains(contType, myhttp.ContentTypeJSON) {
		var err error
		if ret, err = serder.JSONToObjectStreamEx[TBody](resp.Body); err != nil {
			return ret, fmt.Errorf("parsing response: %w", err)
		}

		return ret, nil
	}

	cont, err := io.ReadAll(resp.Body)
	if err != nil {
		return ret, fmt.Errorf("unknow response content type: %s, status: %d", contType, resp.StatusCode)
	}
	strCont := string(cont)

	return ret, fmt.Errorf("unknow response content type: %s, status: %d, body(prefix): %s", contType, resp.StatusCode, strCont[:math2.Min(len(strCont), 200)])
}
