#!/bin/bash
# Compile on Linux:
# sudo apt install clang libasound2-dev
set -e
# get version from main.go VERSION string
if [ $# -eq 0 ]; then
	VERSION=$(cat main.go|grep "VERSION string"| awk -v FS="(\")" '{print $2}')
else
  VERSION=$1
fi
echo Releasing $VERSION
WINF=obs-mcu-$VERSION-windows.zip
LINUXF=obs-mcu-$VERSION-linux.zip
MACF=obs-mcu-$VERSION-macos.zip
RASPIF=obs-mcu-$VERSION-raspberrypi.zip

# build on remote machines
echo "Building Mac Version..."
go build -o obs-mcu
zip $MACF obs-mcu
rm obs-mcu
echo "Building Windows and Linux Version..."
ssh hp-windows -t "cd \\Users\\Normen\\Code\\obs-mcu && git pull && go build && bash -c '~/.go/bin/go build'"
scp hp-windows:~/Code/obs-mcu/obs-mcu.exe ./
zip $WINF obs-mcu.exe
rm obs-mcu.exe
scp hp-windows:~/Code/obs-mcu/obs-mcu ./
zip $LINUXF obs-mcu
rm obs-mcu
echo "Building Raspi Version..."
ssh portapi -t "cd ~/code/obs-mcu && git pull && ~/.go/bin/go build"
scp portapi:~/code/obs-mcu/obs-mcu ./
zip $RASPIF obs-mcu
rm obs-mcu

# publish to github
git pull
set +e
LASTTAG=$(git describe --tags --abbrev=0)
set -e
git log $LASTTAG..HEAD --no-decorate --pretty=format:"- %s" --abbrev-commit > changes.txt
vim changes.txt
gh release create $VERSION $LINUXF $MACF $WINF $RASPIF -F changes.txt -t $VERSION
rm changes.txt
rm *.zip

# update homebrew tap
URL="https://github.com/normen/obs-mcu/archive/$VERSION.tar.gz"
wget -q $URL
SHASUM=$(shasum -a 256 $VERSION.tar.gz|awk '{print$1}')
rm $VERSION.tar.gz
cd ../../BrewCode/homebrew-tap
sed -i bak "s/sha256 \".*/sha256 \"$SHASUM\"/" Formula/obs-mcu.rb
sed -i bak "s!url \".*!url \"$URL\"!" Formula/obs-mcu.rb
rm Formula/obs-mcu.rbbak
git add -A
git commit -m "update obs-mcu to $VERSION"
git push
