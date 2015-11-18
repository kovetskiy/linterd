# Maintainer: Egor Kovetskiy <e.kovetskiy@office.ngs.ru>
pkgname=linterd
pkgver=20151118.3_e9f30aa
pkgrel=1
pkgdesc="lint your golang project with gometalinter"
arch=('i686' 'x86_64')
license=('GPL')
makedepends=('go' 'git')
depends=('go' 'git' 'gometalinter-git')

source=(
	"linterd::git+https://github.com/kovetskiy/linterd"
	"linterd.service"
	"linterd.conf"
)

md5sums=(
	'SKIP'
	'd12911624328cbd6a69677c31b16650e'
	'05eea72c8bee6b30390f03ea49962bd7'
)

backup=(
    "etc/linterd.conf"
)

pkgver() {
	cd "$srcdir/$pkgname"
	local date=$(git log -1 --format="%cd" --date=short | sed s/-//g)
	local count=$(git rev-list --count HEAD)
	local commit=$(git rev-parse --short HEAD)
	echo "$date.${count}_$commit"
}

build() {
	cd "$srcdir/$pkgname"

	if [ -L "$srcdir/$pkgname" ]; then
		rm "$srcdir/$pkgname" -rf
		mv "$srcdir/.go/src/$pkgname/" "$srcdir/$pkgname"
	fi

	rm -rf "$srcdir/.go/src"

	mkdir -p "$srcdir/.go/src"

	export GOPATH="$srcdir/.go"

	mv "$srcdir/$pkgname" "$srcdir/.go/src/"

	cd "$srcdir/.go/src/$pkgname/"
	ln -sf "$srcdir/.go/src/$pkgname/" "$srcdir/$pkgname"

	git submodule init
	git submodule update

	echo "Running 'go get'..."
	GO15VENDOREXPERIMENT=1 go get \
		-ldflags="-X main.main.version=$pkgver-$pkgrel"
}

package() {
	find "$srcdir/.go/bin/" -type f -executable | while read filename; do
		install -DT "$filename" "$pkgdir/usr/bin/$(basename $filename)"
	done
	install -DT -m0755 "$srcdir/linterd.service" "$pkgdir/usr/lib/systemd/system/linterd.service"
	install -DT -m0600 "$srcdir/linterd.conf" "$pkgdir/etc/linterd.conf"
}
