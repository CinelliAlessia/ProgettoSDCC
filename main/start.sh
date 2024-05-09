#!/bin/bash

# Imposta la variabile di ambiente

# 1 locale, 2 docker, 3 EC2
export CONFIG=2
# 1=true 0=false
export DEBUG=0
# 0 = consistenza causale, 1 = consistenza sequenziale
export CONSISTENCY=0
# 0 = test di base, 1 = test medio 0 = test avanzato
export TYPETEST=0

pkill gnome-terminal

if [ "$CONFIG" -eq 1 ]; then

  # Percorsi dei file Go
  server_file="server/"
  client_file="client/"

  # Numero di repliche
  Replicas=3

  # Esecuzione del server in nuovi terminali gnome-terminal
  for ((i=0; i<Replicas; i++)); do
    gnome-terminal --geometry=100x24+0+0 -- bash -c "cd $server_file && go run . $i; exec bash" &
  done

  # Esecuzione del client in un nuovo terminale gnome-terminal
  gnome-terminal --geometry=100x24+0+0 -- bash -c "cd $client_file && go run .; exec bash" &

elif [ "$CONFIG" -eq 2 ]; then
  # Esecuzione con Docker Compose
  docker-compose up --build
elif [ "$CONFIG" -eq 3 ]; then
  echo "Esecuzione su EC2"

  ssh -i "SDCC2324.pem" ec2-user@ec2-35-153-131-38.compute-1.amazonaws.com

  echo "Installazione Docker"
  sudo yum update -y
  sudo yum install -y docker

  echo "Installazione Docker Compose"
  sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
  sudo chmod +x /usr/local/bin/docker-compose

  echo "docker demon"
  sudo service docker start

  echo "Installazione git e Clonazione del repository"
  sudo yum install git -y
  git clone https://github.com/CinelliAlessia/ProgettoSDCC/main

  echo "Run docker compose:"
  sudo docker-compose -f compose.yml up

else
  echo "Configurazione non supportata, inserire 1, 2 o 3."
fi
