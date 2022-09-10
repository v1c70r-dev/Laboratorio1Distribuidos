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

// Permite coenxión cola síncrona proto
type server struct {
	pb.UnimplementedMessageServiceServer
}

func (s *server) Intercambio(ctx context.Context, msg *pb.Message) (*pb.Message, error) {
	fmt.Println(msg.Body)
	return &pb.Message{Body: "NO"}, nil
}

func main() {

	labName := "labRenca"                                           //nombre del laboratorio
	helpQueue := "SOS"                                              //nombre de la cola
	hostQ := "localhost"                                            //ip del servidor de RabbitMQ 172.17.0.1
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

	//Mensaje enviado a la cola de RabbitMQ (Llamado de emergencia)
	body := msgEstallido(labName)
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
	log.Printf(" [x] Sent %s\n", body)

	//Conexión cola síncrona (proto)
	listener, err := net.Listen("tcp", ":50051") //conexion sincrona
	if err != nil {
		panic("La conexion no se pudo crear" + err.Error())
	}

	serv := grpc.NewServer()
	for {
		pb.RegisterMessageServiceServer(serv, &server{})
		if err = serv.Serve(listener); err != nil {
			panic("El server no se pudo iniciar" + err.Error())
		}
	}
}

/*
Retorna el nombre del lab si ocurre un estallido
*/
func msgEstallido(lab_name string) string {
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
		fmt.Println("Analizando estado Laboratorio: [ ESTALLIDO ]")
	} else {
		estallido = false //false
		fmt.Println("Analizando estado Laboratorio: [ OK ]")
	}

	return estallido
}
