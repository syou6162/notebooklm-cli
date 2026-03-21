package main

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/sys/unix"
	"howett.net/plist"
)

const finderCommentAttr = "com.apple.metadata:kMDItemFinderComment"

// GetFinderComment はファイルのFinderコメントを取得する
func GetFinderComment(path string) (string, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("ファイルが存在しません: %s", path)
		}
		return "", fmt.Errorf("ファイルの状態取得エラー: %w", err)
	}

	size, err := unix.Getxattr(path, finderCommentAttr, nil)
	if err != nil {
		if errors.Is(err, unix.ENOATTR) {
			return "", nil
		}
		return "", fmt.Errorf("拡張属性のサイズ取得エラー: %w", err)
	}

	if size == 0 {
		return "", nil
	}

	buf := make([]byte, size)
	n, err := unix.Getxattr(path, finderCommentAttr, buf)
	if err != nil {
		return "", fmt.Errorf("拡張属性の取得エラー: %w", err)
	}

	var comment string
	if _, err := plist.Unmarshal(buf[:n], &comment); err != nil {
		return "", fmt.Errorf("plistのデコードエラー: %w", err)
	}

	return comment, nil
}

// SetFinderComment はファイルのFinderコメントを設定する
func SetFinderComment(path string, comment string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("ファイルが存在しません: %s", path)
		}
		return fmt.Errorf("ファイルの状態取得エラー: %w", err)
	}

	data, err := plist.Marshal(comment, plist.BinaryFormat)
	if err != nil {
		return fmt.Errorf("plistエンコードエラー: %w", err)
	}

	if err := unix.Setxattr(path, finderCommentAttr, data, 0); err != nil {
		return fmt.Errorf("拡張属性の設定エラー: %w", err)
	}

	return nil
}
