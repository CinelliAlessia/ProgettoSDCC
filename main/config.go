package main

// Configurazione globale del programma

// Replicas Numero di repliche del server -> almeno 3
const Replicas = 3

// ReplicaPorts Lista delle porte su cui le repliche possono essere contattate
var ReplicaPorts = [Replicas]string{"8085", "8086", "8087"}
