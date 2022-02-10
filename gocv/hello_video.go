import (
	"gocv.io/x/gocv"
)

func main() {
	// ビデオキャプチャーを開始
	// webcamにはビデオキャプチャで撮ってきたものを入れている
	webcam, _ := gocv.VideoCaptureDevice(0)
	// キャプチャー画面の上部に表示される文字列を設定
	window := gocv.NewWindow("Hello")
	// 映像を画像に変換
	img := gocv.NewMat()

	for {
		// webcamから読み込んだものをimgのMatに格納
		webcam.Read(&img)
		// ウェブカメラから取得した映像をウィンドウに表示
		window.IMShow(img)
		// 指定した時間毎にウィンドウを更新
		window.WaitKey(1)
	}
}
