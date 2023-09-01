package os

import (
	"os"
	"path/filepath"

	"gitlink.org.cn/cloudream/common/pkgs/iterator"
)

type DirIterator struct {
	rootPath    string
	walked      bool
	walkedInfos []FileInfo
	index       int
}

type FileInfo struct {
	Path string
	Info os.FileInfo
}

func (i *DirIterator) MoveNext() (*FileInfo, error) {
	if !i.walked {
		i.walked = true
		// TODO 可以考虑优化成MoveNext一次就产生一个FileInfo的形式
		err := filepath.WalkDir(i.rootPath, func(fname string, fi os.DirEntry, err error) error {
			if err != nil {
				return nil
			}

			if !fi.IsDir() {
				info, err := fi.Info()
				if err != nil {
					return err
				}

				i.walkedInfos = append(i.walkedInfos, FileInfo{
					Path: fname,
					Info: info,
				})
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	if i.index >= len(i.walkedInfos) {
		return nil, iterator.ErrNoMoreItem
	}

	item := i.walkedInfos[i.index]
	i.index++
	return &item, nil
}

func (i *DirIterator) Close() {

}

func WalkDir(rootPath string) *DirIterator {
	return &DirIterator{
		rootPath: rootPath,
	}
}
