# modis
A library to work with MODIS HDF files.

## System requirements

Modis delivers its products in `hdf4` format and this library uses `gdal` 
to operate on those files. As such it requires `gdal` with dev headers
and support for `tiff` and `hdf4`.

### OSX

The standard `homebrew` port comes with `hdf5`, while `hdf4` is generally 
not available in `homebrew`. This necessitates installing `gdal` from 
source. Besides, source installation allows to dramatically reduce the 
number of dependencies.  

1.  Install dependencies available in `homebrew`

    ```
    brew install pkg-config
    brew install libtiff
    brew install libgeotiff
    brew install proj
    brew install gfortran
    ```

1.  Download and install `hdf4`

    Download from 
[hdfgroup.org](https://portal.hdfgroup.org/display/support/Download+HDF4), 
e.g. `hdf-4.2.15.tar.gz`.

    Configure, make and install:
    ```
    CFLAGS=-Wno-implicit-function-declaration ./configure --prefix=/usr/local --enable-shared=yes --disable-static --disable-fortran --with-jpeg=/opt/homebrew/Cellar/jpeg/9e
    make && make install
    ```

1.  Download and install `gdal`

    Download from [gdal.org](https://gdal.org/download.html)

    Configure, make and install:
    ```
    ./configure --prefix=/usr/local --enable-shared=yes --disable-static --with-hdf4=yes --with-proj=/opt/homebrew/Cellar/proj/9.0.0_1 --with-jpeg=/opt/homebrew/Cellar/jpeg/9e --with-libtiff=/opt/homebrew/Cellar/libtiff/4.3.0 --with-geotiff=/opt/homebrew/Cellar/libgeotiff/1.7.1_1
    make && make install
    ```

Now check supported `gdal` formats (`gdal-config --formats`) and build the library `go build .`.


Copyright © 2020 Ekaterina Sklyar & Oleg Sklyar. All rights reserved.
