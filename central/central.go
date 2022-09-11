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

// labRenca, labPohan, labPripiat
var puertos = [4]string{":50051", ":50055", ":50059", ":50063"}
var equipo1_ = true //disponibilidad
var equipo2_ = true

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

// func estadoContencionFun(serviceCliente){
// 	res, err := serviceCliente.ContencionStatus(
// 		context.Background(),
// 		&pb.Message{
// 			Body: "Revisando estado Escuadrón: ",
// 		})
// 	if err != nil {
// 		panic("No se puede crear el mensaje " + err.Error())
// 	}
// 	//log.Println(res.Status, res.Body) //respuesta del laboratorio
// 	//Si esta contenido el estallido, ¿el lab cierra conexión con la central?
// 	if res.Status.String() == "NOLISTO" {
// 		//Checkea estado contencion
// 		log.Println("Status Escuadra X: " + res.Status.String() + " " + res.Body)
// 		time.Sleep(5 * time.Second) //espera de 5 segundos
// 	} else {
// 		log.Println("_Status Escuadra X: " + res.Status.String() + " " + res.Body)
// 		//connS.Close()
// 		if res.Body == "equipo1" {
// 			log.Println("if del equipo1")
// 			equipo1_ = true //disponible
// 		} else if res.Body == "equipo2" {
// 			log.Println("if del equipo2")
// 			equipo2_ = true //disponible
// 		}
// 	}
// }

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
		} else if labName == "labPripiat" {
			port = puertos[2]
		}

		fmt.Println("port: ", port)

		connS, err := grpc.Dial(hostS+port, grpc.WithInsecure()) //crea la conexion sincrona con el laboratorio

		if err != nil {
			panic("No se pudo conectar con el servidor" + err.Error())
		}

		serviceCliente := pb.NewMessageServiceClient(connS)

		/*****************Estado Contencion*********************/

		//Manda info a los lab sobre la disp de los equipos de mercenarios
		resDisp, errDisp := serviceCliente.CheckDispEscuadron(
			context.Background(),
			&pb.Escuadron{
				Equipo1: equipo1_, Equipo2: equipo2_,
			})
		if errDisp != nil {
			panic("No se puede crear el mensaje " + err.Error())
		}

		//Recibe el nombre del equipo elegido por el lab. El equipo elegido
		//cambia su valor a false ¿pero q lab?
		if resDisp.Equipox == "equipo1" {
			equipo1_ = false //Equipo ocupa contenienndo a los bichos
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
				//connS.Close()
				if res.Body == "equipo1" {
					log.Println("if del equipo1")
					equipo1_ = true //disponible
				}

				if res.Body == "equipo2" {
					log.Println("if del equipo2")
					equipo2_ = true //disponible
				}
			}
		} else if resDisp.Equipox == "equipo2" {
			equipo2_ = false
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
				//connS.Close()
				if res.Body == "equipo1" {
					log.Println("if del equipo1")
					equipo1_ = true //disponible
				}

				if res.Body == "equipo2" {
					log.Println("if del equipo2")
					equipo2_ = true //disponible
				}
			}
		}

		/*****************Estado Contencion*********************/
		// log.Println("equpo1_", equipo1_)
		// if !equipo1_ {

		// 	res, err := serviceCliente.ContencionStatus(
		// 		context.Background(),
		// 		&pb.Message{
		// 			Body: "Revisando estado Escuadrón: ",
		// 		})
		// 	if err != nil {
		// 		panic("No se puede crear el mensaje " + err.Error())
		// 	}
		// 	//log.Println(res.Status, res.Body) //respuesta del laboratorio
		// 	//Si esta contenido el estallido, ¿el lab cierra conexión con la central?
		// 	if res.Status.String() == "NOLISTO" {
		// 		//Checkea estado contencion
		// 		log.Println("Status Escuadra X: " + res.Status.String() + " " + res.Body)
		// 		time.Sleep(5 * time.Second) //espera de 5 segundos
		// 	} else {
		// 		log.Println("_Status Escuadra X: " + res.Status.String() + " " + res.Body)
		// 		//connS.Close()
		// 		if res.Body == "equipo1" {
		// 			log.Println("if del equipo1")
		// 			equipo1_ = true //disponible
		// 		} else if res.Body == "equipo2" {
		// 			log.Println("if del equipo2")
		// 			equipo2_ = true //disponible
		// 		}
		// 	}
		// }

		// log.Println("equpo2_", equipo2_)
		// if !equipo2_ {

		// 	res, err := serviceCliente.ContencionStatus(
		// 		context.Background(),
		// 		&pb.Message{
		// 			Body: "Revisando estado Escuadrón: ",
		// 		})
		// 	if err != nil {
		// 		panic("No se puede crear el mensaje " + err.Error())
		// 	}
		// 	//log.Println(res.Status, res.Body) //respuesta del laboratorio
		// 	//Si esta contenido el estallido, ¿el lab cierra conexión con la central?
		// 	if res.Status.String() == "NOLISTO" {
		// 		//Checkea estado contencion
		// 		log.Println("Status Escuadra X: " + res.Status.String() + " " + res.Body)
		// 		time.Sleep(5 * time.Second) //espera de 5 segundos
		// 	} else {
		// 		log.Println("_Status Escuadra X: " + res.Status.String() + " " + res.Body)
		// 		//connS.Close()
		// 		if res.Body == "equipo1" {
		// 			//log.Println("if del equipo1")
		// 			equipo1_ = true //disponible
		// 		} else if res.Body == "equipo2" {
		// 			//log.Println("if del equipo2")
		// 			equipo2_ = true //disponible
		// 		}
		// 	}
		// }

	}
	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
