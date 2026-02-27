cask "quramate" do
  version "1.1.3"
  sha256 :no_check

  url "https://github.com/RealDewKJ/QuraMate/releases/download/v#{version}/QuraMate-macOS-universal.zip"
  name "QuraMate"
  desc "Modern, open-source database management tool"
  homepage "https://github.com/RealDewKJ/QuraMate"

  livecheck do
    url :url
    strategy :github_latest
  end

  app "QuraMate.app"

  zap trash: [
    "~/Library/Application Support/QuraMate",
    "~/Library/Preferences/com.wails.QuraMate.plist",
    "~/Library/Caches/QuraMate",
  ]
end
