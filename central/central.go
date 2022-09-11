package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"time"

	pb "Laboratorio1Distribuidos/proto"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
)

// labRenca, labPohang, labPripiat, labKampala
var (
	puertos  = [4]string{":50051", ":50055", ":50059", ":50063"}
	equipo1_ = true //disponibilidad
	equipo2_ = true
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
	port := "" //puerto de la conexion con el laboratorio

	for delivery := range chDelivery {

		labName := string(delivery.Body)
		fmt.Println("Mensaje asíncrono de laboratorio " + labName + " leído") //obtiene el primer mensaje de la cola

		if labName == "labRenca" {
			port = puertos[0]
		} else if labName == "labPohang" {
			port = puertos[1]
		} else if labName == "labPripiat" {
			port = puertos[2]
		} else if labName == "labKampala" {
			port = puertos[3]
		}
		/** Crea la conexion sincrona con el laboratorio **/
		connS, err := grpc.Dial(hostS+port, grpc.WithInsecure())

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

		//Recibe el nombre del equipo elegido y el nombre del lab. El equipo elegido
		//cambia su valor a false => equipo está ocupado

		go func() {

			if resDisp.Equipox == "Escuadra1" {
				equipo1_ = false
				primeraLlegada := true
				escuadronNoListo := true
				log.Println("Se envía " + resDisp.Equipox + " a Laboratorio " + resDisp.NombreLab)

				for escuadronNoListo {
					//Se manda el escuadron al lab
					res, err := serviceCliente.ContencionStatus(
						context.Background(),
						&pb.EquipoEnviadoPorCentral{
							Eepc:           resDisp.Equipox, //se manda el equipo al laboratorio
							PrimeraLlegada: primeraLlegada,
						})
					if err != nil {
						panic("No se puede crear el mensaje " + err.Error())
					}

					//Se recibe el estado de contencion y nombre del escuadron
					//Si el estallido está contenido, se cierra la conexión con el lab
					if res.Status.String() == "NOLISTO" {
						log.Println("Status " + res.NombreEscuadron + ": " + res.Status.String())
						time.Sleep(5 * time.Second) //espera de 5 segundos
						primeraLlegada = false
					} else {
						escuadronNoListo = false
						log.Println("Status " + res.NombreEscuadron + ": " + res.Status.String())
						log.Println("Retorno a Central " + res.NombreEscuadron + ", Conexión Laboratorio" + resDisp.NombreLab + "Cerrada")
						equipo1_ = true //vuelve a quedar disponible
						connS.Close()   //Se cierra la conexión
					}
				}

			}
		}()

		go func() {
			if resDisp.Equipox == "Escuadra2" {
				equipo2_ = false
				primeraLlegada := true
				escuadronNoListo := true
				log.Println("Se envía " + resDisp.Equipox + " a Laboratorio " + resDisp.NombreLab)

				for escuadronNoListo {
					//Se manda el escuadron al lab
					res, err := serviceCliente.ContencionStatus(
						context.Background(),
						&pb.EquipoEnviadoPorCentral{
							Eepc:           resDisp.Equipox, //se manda el equipo al laboratorio
							PrimeraLlegada: primeraLlegada,
						})
					if err != nil {
						panic("No se puede crear el mensaje " + err.Error())
					}

					//Se recibe el estado de contencion y nombre del escuadron
					//Si el estallido está contenido, se cierra la conexión con el lab
					if res.Status.String() == "NOLISTO" {
						log.Println("Status " + res.NombreEscuadron + ": " + res.Status.String() + resDisp.Equipox)
						time.Sleep(5 * time.Second) //espera de 5 segundos
						primeraLlegada = false
					} else {
						escuadronNoListo = false
						log.Println("Status " + res.NombreEscuadron + ": " + res.Status.String())
						log.Println("Retorno a Central " + res.NombreEscuadron + ", Conexión Laboratorio" + resDisp.NombreLab + "Cerrada")
						equipo2_ = true //vuelve a quedar disponible
						connS.Close()   //Se cierra la conexión
					}
				}
			}
		}()

		// else if resDisp.Equipox == "NOHAYESCUADRA" {
		// 	/*La idea es q Solo se consuman msg de rabbit cuando exista al menos un equipo
		// 	disponible en la central*/
		// 	log.Println("NOHAYESCUADRA")
		// }

		// go func() {
		// 	// Fin ejecucion programa
		// 	// Capturar ctrl + c
		// 	c := make(chan os.Signal, 1)
		// 	signal.Notify(c, os.Interrupt)

		// 	for sig := range c {

		// 		log.Println(sig)
		// 		// Proto mande msg manera sincrona a todos los labs
		// 		resFin, errFin := serviceCliente.FinPrograma(
		// 			context.Background(),
		// 			&pb.MessageTermino{
		// 				EndSignal: true,
		// 				MsgFin:    "Lab termine su ejecucion",
		// 			})
		// 		if errFin != nil {
		// 			panic("No se puede crear el mensaje " + err.Error())
		// 		}

		// 		//lab envian señal de vuelta -> 4 señales antes de morir
		// 		log.Println(resFin.MsgFin)
		// 		// sig is a ^C, handle it
		// 		time.Sleep(1 * time.Second)
		// 		//os.Exit(1)
		// 	}

		// 	// lab confirman
		// 	// se terminan de ejecutar

		// }()

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			for sig := range c {
				// sig is a ^C, handle it
				// Proto mande msg manera sincrona a todos los labs
				// resFin, errFin := serviceCliente.FinPrograma(
				// 	context.Background(),
				// 	&pb.MessageTermino{
				// 		EndSignal: true,
				// 		MsgFin:    "Lab termine su ejecucion",
				// 	})
				// if errFin != nil {
				// 	panic("No se puede crear el mensaje " + err.Error())
				// }
				// //lab envian señal de vuelta -> 4 señales antes de morir
				// log.Println(resFin.MsgFin)
				// log.Println(sig)
				// if resFin.EndSignal {
				// 	log.Println("SEACABO!!")
				// 	//os.Exit(1)
				// }
				log.Println(sig)
			}
		}()
	}

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
