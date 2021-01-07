FROM debian:jessie-slim

workdir /root/
COPY . .

RUN apt update -y && apt install openvpn -y && apt install curl -y