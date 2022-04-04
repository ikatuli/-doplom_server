# docker build --tag doplom_server ./ 
FROM archlinux
RUN pacman-db-upgrade
RUN pacman -Suy  gcc-go --noconfirm
RUN yes | pacman -Scc 

RUN mkdir /app
WORKDIR /app
COPY main.go ./
COPY go.mod ./
COPY static/ ./static
COPY user/ ./user
RUN go mod download
RUN go mod tidy
RUN go build -compiler=gccgo main.go

EXPOSE  9000

CMD [ "/app/main" ]

# docker run -p 9000:9000 doplom_server
