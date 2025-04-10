class Kiosk < Formula
  desc "Terminal workspace manager"
  homepage "https://github.com/jmelowry/kiosk"
  version "main-1744261656"
  if OS.mac?
    depends_on "go" => :build

    def install
      system "go", "build", *std_go_args(output: bin/"kiosk")
    end
  else
    url "https://github.com/jmelowry/kiosk/releases/download/main-1744261656/kiosk-main-1744261656-linux-amd64.tar.gz"
    sha256 "62ee901b2336e6595ddf8587a73515abf48eb7ad76831cfab340901ad7401b3a"

    def install
      bin.install "kiosk"
    end
  end

  test do
    system "#{bin}/kiosk", "--help"
  end
end