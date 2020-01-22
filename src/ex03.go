package main

import (
    "fmt"
    "sync"
)

// la fonction Max retourne la valeur la plus grande entre les deux paramètres fournis
func Max(x, y int) int {
    if x < y {
        return y
    }
    return x
}

// définition d'une structure de Message
type Message struct {
    id int // l'id d'un noeud
    isLeader bool // si l'id envoyé est considéré comme un leader
}

// définition d'une structure de Noeud
type RingNode struct {
    inChannel <-chan Message // la connexion entrante
    outChannel chan<- Message // la connexion sortante
    localId int // l'id du noeud
    leaderId int // l'id du leader
}

func NewRingNode(id int, inChannel <-chan Message, outChannel chan<- Message) *RingNode {
    return &RingNode{inChannel, outChannel, id, 0}
}

// code principal des noeuds
func (r *RingNode) Run(wg *sync.WaitGroup) {
	// Done va etre executée un fois la boucle terminée (le inChannel est fermée)
	defer wg.Done()

    // Initialisation: chaque Noeud envoie son id au suivant
    fmt.Println("From", r.localId, ": Sending ", r.localId)
    r.outChannel <- Message{r.localId, false}

    for receivedMessage := range r.inChannel {
        fmt.Println("From", r.localId, ": received", receivedMessage.id, receivedMessage.isLeader)

		if r.leaderId != 0 {
			// si le leaderId à été modifié alors le leader à été identifié, on ne fait rien
			continue
		} else if !receivedMessage.isLeader { // si le leader n'as pas encore été idenfié
			
			// si l'id recu est l'id du noeud courant
			if receivedMessage.id == r.localId {
				// si l'id recu est égal à celui du noeud courant alors on viens d'identifier le leader
				r.leaderId = receivedMessage.id // on met à jour la valeur de l'id du leader
				fmt.Println("From", r.localId, ": Leader found at", r.leaderId)
				r.outChannel <- Message{r.leaderId, true} // on envoie un message contenant l'id du leader et on met à jour la valeur isLeader à true
				close(r.outChannel) // on ferme le channel
			} else { // si l'id recu n'est pas l'id courant
				max := Max(r.localId, receivedMessage.id) // on compare les deux id et on renvoie le plus grand des deux
				fmt.Println("From", r.localId, ": compared (", r.localId, "-", receivedMessage.id, ") | sending:", max) // on affiche un message informatif
				r.outChannel <- Message{max, false} // on renvoie le plus grand id des deux
			}
		
		} else if receivedMessage.isLeader { // si le leader est identifié (peut etre remplacé par un else standard)
		
			r.leaderId = receivedMessage.id // on met à jour la valeur de leaderId avec l'id recu
			fmt.Println("From", r.localId, ": Leader found at", r.leaderId, "| spreading message and closing channel")
			r.outChannel <- receivedMessage // on envoie le message recu au noeud suivant (pas besoin de le modifier)
			close(r.outChannel) // on ferme la connexion sortante du noeud courant

		}
	}
}

func main() {

	const N = 200 // défini le nombre de noeuds à créer sur l'anneau
	var outChannels [N]chan Message  // une slice permettant de stocker les channels utilisés par les noeud de l'anneau
	var nodes [N]*RingNode // une slice permettant de stocker les noeud de l'anneau

	var wg sync.WaitGroup
    wg.Add(N)

	// Création des channels
	for i := range outChannels {
		outChannels[i] = make(chan Message, 1)
	}

	// Création des noeuds
	for i := 0; i < N; i++ {
		if i == 0 { // la création du premier noeud utilise le dernier chan en entrée et le premier chan comme sortie (pour fermer l'anneau)
			nodes[i] = NewRingNode(i + 1, outChannels[N - 1], outChannels[i]) // on ajoute le noeud créé précédement créé
		} else {
			nodes[i] = NewRingNode(i + 1, outChannels[i-1], outChannels[i]) // on instancie une instance de RingNode et y lie la channel précédente et la suivante
		}	

		go nodes[i].Run(&wg)
	}

	fmt.Println("Initialisation of", N, "node(s)")

    wg.Wait()
}