DIST ?= $(shell rpm --eval %{dist})
SPECFILE ?= golbd.spec
 
PKG ?= $(shell rpm -q --specfile $(SPECFILE) --queryformat "%{name}-%{version}\n" | head -n 1)
RPMBUILD_ARG = --define 'dist $(DIST)' --define "_topdir $(PWD)/build" --define '_sourcedir $(PWD)/SOURCES' $(SPECFILE)

srpm:
	echo "Creating the source rpm"
	mkdir -p SOURCES version
	go mod vendor
	tar zcf SOURCES/$(PKG).tgz  --exclude SOURCES --exclude .git --exclude .koji --exclude .gitlab-ci.yml --transform "s||$(PKG)/|" .
	rpmbuild -bs $(RPMBUILD_ARG)
   
rpm: srpm
	echo "Creating the rpm"
	rpmbuild -bb $(RPMBUILD_ARG)

clean:
	rm -rf build vendor
