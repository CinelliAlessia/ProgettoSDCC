#!/bin/bash

# Leggi il numero di repliche dalla variabile d'ambiente
REPLICATION_SERVER="$REPLICATION_SERVER"

# Sostituisci il segnaposto nel file docker-compose.yml con il numero di repliche
sed -i "s/REPLICA_PLACEHOLDER/$REPLICATION_SERVER/" docker-compose.yml
