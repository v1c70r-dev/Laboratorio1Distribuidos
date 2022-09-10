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

	for delivery := range chDelivery {
		port := ":50051"                                           //puerto de la conexion con el laboratorio
		fmt.Println("Pedido de ayuda de " + string(delivery.Body)) //obtiene el primer mensaje de la cola
		connS, err := grpc.Dial(hostS+port, grpc.WithInsecure())   //crea la conexion sincrona con el laboratorio

		if err != nil {
			panic("No se pudo conectar con el servidor" + err.Error())
		}

		serviceCliente := pb.NewMessageServiceClient(connS)

		for {
			//envia el mensaje al laboratorio
			// res, err := serviceCliente.Intercambio(context.Background(),
			// 	&pb.Message{
			// 		Body: "Revisando estado Escuadrón: ",
			// 	})

			// if err != nil {
			// 	panic("No se puede crear el mensaje " + err.Error())
			// }

			// log.Println(res.Body)       //respuesta del laboratorio
			// time.Sleep(5 * time.Second) //espera de 5 segundos
			//defer connS.Close()

			//Revisa el estado de contencion/escuadron del laboratorio
			//enviando un mensaje al laboratorio
			res, err := serviceCliente.ContencionStatus(
				context.Background(),
				&pb.Message{
					Body: "Revisando estado Escuadrón: ",
				})

			if err != nil {
				panic("No se puede crear el mensaje " + err.Error())
			}

			log.Println(res.Status, res.Body) //respuesta del laboratorio
			// time.Sleep(5 * time.Second)       //espera de 5 segundos
			// defer connS.Close()

			//Si esta contenido el estallido, se cierra la conexion con el lab
			if res.Status.String() == "LISTO" {
				log.Println("Laboratorio ha cerrado la conexión")
			} else {
				//Continua checkeando el estado de la contencion
				time.Sleep(5 * time.Second) //espera de 5 segundos
			}
		}
	}

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
