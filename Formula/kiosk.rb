class Kiosk < Formula
  desc "Terminal workspace manager"
  homepage "https://github.com/jmelowry/kiosk"
  version "main-1744334106"
  if OS.mac?
    depends_on "go" => :build

    def install
      system "go", "build", *std_go_args(output: bin/"kiosk")
    end
  else
    url ""
    sha256 ""

    def install
      bin.install "kiosk"
    end
  end

  test do
    system "#{bin}/kiosk", "--help"
  end
end