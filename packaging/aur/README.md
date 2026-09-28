# Publishing to the AUR

This directory is the complete contents of the separate `gftp` AUR Git
repository. It is not pushed to the upstream source repository.

After creating an AUR account and configuring its SSH key, publish it with:

```sh
git clone ssh://aur@aur.archlinux.org/gftp.git
cp packaging/aur/PKGBUILD packaging/aur/.SRCINFO gftp/
cd gftp
makepkg -si
namcap PKGBUILD *.pkg.tar.zst
git add PKGBUILD .SRCINFO
git commit -m 'feat: initial gftp package'
git push
```

The package currently builds the published `v0.1.0` release. The `LICENSE`
file was added after that release, so create a new upstream release before the
first AUR publication if you want it included in the packaged documentation.
For a new release, update `pkgver`, run `updpkgsums`, regenerate `.SRCINFO`
with `makepkg --printsrcinfo > .SRCINFO`, test with `makepkg -si`, and push.
