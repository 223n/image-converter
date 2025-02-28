/*
Package converter の一部として、AVIF変換に特化した関数を提供します。
*/
package converter

import (
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"

	"github.com/Kagami/go-avif"
	"github.com/yourusername/image-converter/internal/config"
)

// SaveAVIF は画像をAVIFとして保存します
func SaveAVIF(img image.Image, outputPath string) error {
	output, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer output.Close()

	// AVIFエンコードオプションの設定
	options := prepareAVIFOptions()

	// AVIF形式で保存
	log.Printf("AVIF変換開始: %s (品質: %d, 速度: %d)",
		outputPath, options.Quality, options.Speed)

	if err := avif.Encode(output, img, options); err != nil {
		return err
	}

	// エンコード後のファイルサイズを確認
	fi, err := os.Stat(outputPath)
	if err != nil || fi.Size() == 0 {
		return fmt.Errorf("AVIF変換に失敗しました: 出力ファイルサイズが0バイトです")
	}

	log.Printf("AVIF変換完了: %s (サイズ: %d バイト)", outputPath, fi.Size())
	return nil
}

// prepareAVIFOptions はAVIF変換オプションを準備します
func prepareAVIFOptions() *avif.Options {
	options := &avif.Options{}

	// Quality: 品質 (0-100)
	// go-avifライブラリでは1-63の範囲の値が有効
	quality := config.GetAVIFQuality()
	if quality > 63 {
		log.Printf("警告: AVIF品質値が範囲外です。63に調整します: %d -> 63", quality)
		options.Quality = 63
	} else if quality < 1 {
		log.Printf("警告: AVIF品質値が範囲外です。1に調整します: %d -> 1", quality)
		options.Quality = 1
	} else {
		options.Quality = quality
	}

	// Speed: 処理速度 (0-10, 値が大きいほど速いが品質は下がる)
	// go-avifライブラリでは0-10の範囲の値が有効
	speed := config.GetAVIFSpeed()
	if speed > 10 {
		log.Printf("警告: AVIF速度値が範囲外です。10に調整します: %d -> 10", speed)
		options.Speed = 10
	} else if speed < 0 {
		log.Printf("警告: AVIF速度値が範囲外です。0に調整します: %d -> 0", speed)
		options.Speed = 0
	} else {
		options.Speed = speed
	}

	return options
}

// ConvertToAVIF は公開APIとして高レベルのAVIF変換機能を提供します
func ConvertToAVIF(img image.Image, outputPath string) error {
	// パス関連の処理
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("出力ディレクトリの作成に失敗しました: %v", err)
	}

	// 実際の変換処理
	if err := SaveAVIF(img, outputPath); err != nil {
		return fmt.Errorf("AVIF変換に失敗しました: %v", err)
	}

	return nil
}

// IsAVIFSupported はAVIFがサポートされているかどうかを確認します
func IsAVIFSupported() bool {
	// libaomの存在を確認する簡易的な方法
	// 実際には複数の方法で確認する必要があるかもしれません
	_, err := os.Stat("/usr/lib/libaom.so")
	if err == nil {
		return true
	}

	_, err = os.Stat("/usr/lib/x86_64-linux-gnu/libaom.so")
	if err == nil {
		return true
	}

	// macOSの場合
	_, err = os.Stat("/usr/local/lib/libaom.dylib")
	if err == nil {
		return true
	}

	return false
}

// GetAVIFInfo はAVIF変換のサポート状況に関する情報を返します
func GetAVIFInfo() string {
	if IsAVIFSupported() {
		return "AVIF変換はサポートされています"
	}

	return "AVIF変換はサポートされていません。libaomをインストールしてください"
}
