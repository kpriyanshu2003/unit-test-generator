# Base image
FROM ubuntu:22.04

# Set timezone
ENV TZ=UTC
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Install dependencies
RUN apt-get update -yqq && apt-get install -yqq --no-install-recommends \
    software-properties-common \
    sudo curl wget cmake make pkg-config locales git \
    gcc-11 g++-11 openssl libssl-dev libjsoncpp-dev uuid-dev \
    zlib1g-dev libc-ares-dev postgresql-server-dev-all \
    libmariadb-dev libsqlite3-dev libhiredis-dev \
    lcov gcovr \
    && rm -rf /var/lib/apt/lists/* \
    && locale-gen en_US.UTF-8

# Set environment variables for localization and compiler
ENV LANG=en_US.UTF-8 \
    LANGUAGE=en_US:en \
    LC_ALL=en_US.UTF-8 \
    CC=gcc-11 \
    CXX=g++-11 \
    AR=gcc-ar-11 \
    RANLIB=gcc-ranlib-11 \
    IROOT=/install

# Install GTest for unit testing
RUN apt-get update && apt-get install -y libgtest-dev cmake
RUN cd /usr/src/gtest && cmake . && make && cp lib/*.a /usr/lib

# Clone Drogon
ENV DROGON_ROOT="$IROOT/drogon"
RUN git clone https://github.com/drogonframework/drogon $DROGON_ROOT

# Build Drogon
WORKDIR $DROGON_ROOT
RUN ./build.sh

# Copy your project source
WORKDIR /app
COPY . /app

# Build project with coverage flags
RUN mkdir -p build && cd build && \
    cmake -DCMAKE_CXX_FLAGS="--coverage" -DCMAKE_EXE_LINKER_FLAGS="--coverage" .. && \
    make VERBOSE=1


# Default command (can be overridden by docker-compose)
CMD ["./build/org_chart"]
