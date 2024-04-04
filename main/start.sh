#!/bin/bash

# Imposta la variabile di ambiente
export CONFIG=1 # 1 locale, 2 docker

if [ "$CONFIG" -eq 1 ]; then
  # Percorsi dei file Go
  server_file="server/"
  client_file="client/"

  # Esegui il server in un nuovo terminale gnome-terminal
  gnome-terminal -- bash -c "cd $server_file && go run . 0; exec bash" &
  gnome-terminal -- bash -c "cd $server_file && go run . 1; exec bash" &
  gnome-terminal -- bash -c "cd $server_file && go run . 2; exec bash" &

  # Esegui il client in un nuovo terminale gnome-terminal
  gnome-terminal -- bash -c "cd $client_file && go run .; exec bash" &
elif [ "$CONFIG" -eq 2 ]; then
  # Esegui Docker Compose
  docker-compose up --build
else
  echo "Configurazione non supportata."
fi
