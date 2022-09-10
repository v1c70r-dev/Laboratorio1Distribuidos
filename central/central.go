package main

import (
	"context"
	"fmt"
	"log"

	"time"

	pb "Laboratorio1Distribuidos/proto"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
)

var puertos = [4]string{":50051", ":50055", ":50059", ":50063"} //labRenca, labPpohan,

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {

	helpQueue := "SOS"                                              //Nombre de la cola
	hostQ := "localhost"                                            //Host de RabbitMQ 172.17.0.1
	hostS := "localhost"                                            //Host de un Laboratorio
	conn, err := amqp.Dial("amqp://guest:guest@" + hostQ + ":5672") //Conexion con RabbitMQ
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

	chDelivery, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	var forever chan struct{}
	port := "" //puerto de la conexion con el laboratorio

	for delivery := range chDelivery {

		labName := string(delivery.Body)
		fmt.Println("Pedido de ayuda de " + labName) //obtiene el primer mensaje de la cola

		if labName == "labRenca" {
			port = puertos[0]
		} else if labName == "labPohang" {
			port = puertos[1]
		}

		fmt.Println("port: ", port)

		connS, err := grpc.Dial(hostS+port, grpc.WithInsecure()) //crea la conexion sincrona con el laboratorio

		if err != nil {
			panic("No se pudo conectar con el servidor" + err.Error())
		}

		serviceCliente := pb.NewMessageServiceClient(connS)

		res, err := serviceCliente.ContencionStatus(
			context.Background(),
			&pb.Message{
				Body: "Revisando estado Escuadrón: ",
			})

		if err != nil {
			panic("No se puede crear el mensaje " + err.Error())
		}

		//log.Println(res.Status, res.Body) //respuesta del laboratorio

		//Si esta contenido el estallido, ¿el lab cierra conexión con la central?
		if res.Status.String() == "NOLISTO" {
			//Checkea estado contencion
			log.Println("Status Escuadra X: " + res.Status.String() + " " + res.Body)
			time.Sleep(5 * time.Second) //espera de 5 segundos
		} else {
			log.Println("_Status Escuadra X: " + res.Status.String() + " " + res.Body)
			connS.Close()
		}

	}
	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
