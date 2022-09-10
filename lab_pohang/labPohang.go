package main

import (
	pb "Laboratorio1Distribuidos/proto"
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

// Permite conexión cola síncrona proto
type server struct {
	pb.UnimplementedMessageServiceServer
}

func (s *server) ContencionStatus(ctx context.Context, msg *pb.Message) (*pb.Contencion, error) {
	ecs := pb.Contencion_NOLISTO
	puerto := "50055"
	estadoContencion := ""
	if contencion() {
		estadoContencion = "[LISTO]"
		ecs = pb.Contencion_LISTO
		//se cierra la conexión
		//CLOSE_SYNCHRONOUS_CONNECTION = true
	} else {
		estadoContencion = "[NO LISTO]"
		ecs = pb.Contencion_NOLISTO
	}
	log.Println(msg.Body + estadoContencion)
	return &pb.Contencion{Status: ecs, Body: puerto}, nil
}

// var global
//var grpcServer *grpc.Server

func main() {

	labName := "labPohang" //nombre del laboratorio
	helpQueue := "SOS"     //nombre de la cola
	hostQ := "localhost"   //ip del servidor de RabbitMQ 172.17.0.1

	/******************Conexión cola síncrona (proto)******************/
	go func() {
		listener, err := net.Listen("tcp", ":50055") //conexion sincrona
		if err != nil {
			panic("La conexion no se pudo crear" + err.Error())
		}
		grpcServer := grpc.NewServer()
		pb.RegisterMessageServiceServer(grpcServer, &server{})
		if err = grpcServer.Serve(listener); err != nil {
			panic("El server no se pudo iniciar" + err.Error())
		}
	}()
	time.Sleep(1 * time.Second)

	/******************Conexión con Rabbitmq******************/

	conn, err := amqp.Dial("amqp://guest:guest@" + hostQ + ":5672") //conexion con RabbitMQ

	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		helpQueue, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare a queue")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for {
		//Mensaje enviado a la cola de RabbitMQ (Llamado de emergencia)
		body := Estallido(labName)
		err = ch.PublishWithContext(ctx,
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(body),
			})
		failOnError(err, "Failed to publish a message")
	}
}

/*
Retorna el nombre del lab si ocurre un estallido
*/
func Estallido(lab_name string) string {
	/*Cada 5 segundos, con una probabilidad de 0.8 se verá la ocurrencia de un estallido
	(0.8 ocurre estallido / 0.2 no ocurre estallido)*/
	var estadoEstallido bool
	c := time.Tick(5 * time.Second)

	for range c {
		estadoEstallido = statusUpdateEstallido()
		if estadoEstallido {
			fmt.Println("SOS Enviado a Central. Esperando respuesta...")
			return lab_name
		}
	}
	return "false"
}

func statusUpdateEstallido() bool {
	//Se generan números equiprobables del 1 al 5. Números del 1 al 4 => estallido
	// y 5 => no estallido. Esto da 4/5 (o 0.8) de pbb para estallido y 0.2 de pbb para no estallido.
	estallido := false
	max := 6
	min := 1
	random := rand.Intn(max-min) + min

	if random < 5 {
		estallido = true //true
		fmt.Println("Analizando estado Laboratorio: [ESTALLIDO]")
	} else {
		estallido = false //false
		fmt.Println("Analizando estado Laboratorio: [OK]")
	}

	return estallido
}

/*
Retorna true si el estallido ha sido contenido, false caso contrario
*/
func contencion() bool {
	//Se generan números equiprobables del 1 al 5. Números del 1 al 3 => contenida
	// y 4,5 => no contenida. Esto da 3/5 (o 0.6) de pbb de contención resuelta y 0.4 no
	contencion := false
	max := 6
	min := 1
	random := rand.Intn(max-min) + min
	if random < 4 {
		contencion = true
		// log.Println("Revisando estado Escuadrón: [ LISTO ]")
	} else {
		contencion = false
		// log.Println("Revisando estado Escuadrón: [ NO LISTO ]")
	}

	return contencion
}
