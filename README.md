# Informatique Répartie - TP01

L’objectif de cette séance est de reprendre l’exercice sur l’anneau à jeton de la première semaine et
d’utiliser un principe similaire pour résoudre le problème de l’élection de leader sur un anneau.
<br></br>
Le principe est simple :

- On dispose d’un anneau de **N** sommets (**N** est inconnu de tous)
- Chaque sommet a un identifiant (entre 1 et **N**)
- Un sommet ne peut envoyer un message qu’au sommet suivant et recevoir un message que
du sommet précédent.

L’objectif est qu’à la fin de l’algorithme chaque sommet connaisse l’identifiant le plus élevé sur
l’anneau (n dans ce cas). Le sommet ayant cet identifiant est le leader.
<br></br>
L’algorithme est décomposé en deux phases : avant que le leader ne soit connu et après. Avant que
le leader ne soit connu :

- Initialement chaque sommet envoi son identifiant au sommet suivant dans l’anneau.
- A chaque réception de message un sommet va renvoyer le max entre son identifiant et
l’identifiant reçu.
- Si un sommet reçoit son propre identifiant c’est que son identifiant a fait tout le tour de l’anneau et était à chaque fois le max. Il est donc le leader.

Une fois le leader auto-identifié :

- Il doit envoyer un message "je suis le leader" au suivant sur l’anneau.
- Chaque sommet va propager le message du leader et l’algorithme se termine quand tous les
sommets savent qui est le leader.

Les deux différences avec l’exercice de la semaine 1 sont que le message envoyé contient une
information (c’est n’est pas juste un jeton) et surtout qu’il y a plusieurs messages en circulation au
même moment.

## Première Partie

1. A quel moment un sommet peut fermer sa connexion sortante ?
   
Un sommet peut fermer sa connexion sortante au moment ou il à envoyé le message indiquant l'identifiant du leader (si il ferme au moment de la réception de l'identifiant du leader, cela bloquerai les sommets suivants en attente du message).

2. Complétez le code fournit pour exécuter l'algorithme sur un anneau à trois sommets (seul le contenu de la boucle doit etre modifiée). Faire un affichage à chaque réception de message et afficher le nombre de messages envoyés.

Si on exécute le code fournit, on se rend compte que l'initialisation des Noeud s'effectue correctement (chaque noeud envoie son id à son voisin), et que les message sont correctement recus (les noeud recoivent les id des noeuds les précédant dans l'anneau). Seulement une fois ces deux étapes effectuées, les goroutines s'attendent mutuellement et créé un [deadlock](https://fr.wikipedia.org/wiki/Interblocage).
<br></br>
Afin d'éviter le deadlock, on doit permettre à chaque noeud de comparer l'identifiant recu et son propre id afin de renvoyer l'id le plus élevé. Pour cela on va compléter la boucle `for` présente dans la fonction `Run`.

```go
for receivedMessage := range r.inChannel {
    fmt.Println("From", r.localId, ": received", receivedMessage.id, receivedMessage.isLeader)

    if r.leaderId != 0 {
        // si le leaderId à été modifié alors le leader à été identifié, on ne fait rien
        continue
    } else if receivedMessage.isLeader == false { // si le leader n'as pas encore été idenfié
        
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
    
    } else if receivedMessage.isLeader == true { // si le leader est identifié (peut etre remplacé par un else standard)
        
        // si l'id recu est égal à celui du noeud courant
        if receivedMessage.id != r.localId { // si l'id recu n'est pas l'id courant
            r.leaderId = receivedMessage.id // on met à jour la valeur de leaderId avec l'id recu
            fmt.Println("From", r.localId, ": Leader found at", r.leaderId, "| spreading message and closing channel")
            r.outChannel <- receivedMessage // on envoie le message recu au noeud suivant (pas besoin de le modifier)
            close(r.outChannel) // on ferme la connexion sortante du noeud courant
        }

    }
}
```

3. Généraliser tout le code pour créer un anneau avec **N** sommets (seul le main doit
être modifié). Vous pouvez tester des solutions dans lesquelles l’identifiant des sommets est
différent de leur position sur l’anneau.

Pour cela on va utiliser trois nouvelles variables, à savoir `outChannels` (array qui permet de stocker les channels utilisés par les noeuds), `nodes` (array qui permet de stocker  les noeud), et `N` (le nombre de noeuds présents sur l'anneau). On va ensuite utiliser une boucle `for` afin d'intialiser les noeuds et lancer les goroutines correspondantes.

Cependant il ne faut pas oublier de lier le premier noeud au dernier et inversement, sinon on n'obtient pas un "anneau".

```go
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
```

4. Analyser le nombre de messages envoyés par chaque sommet, quelle solution
pourriez-vous proposer pour réduire le nombre de messages envoyés ? Vous pouvez
modifier les structures et le code selon vos besoins.

Afin de réduire le nombre de message, on va simplement empecher les noeuds d'envoyer un message si l'id recu est plus petit que le plus grand id déjà envoyé. Pour cela on va utiliser un nouvel attribut à `RingNode`, `maxId` qui permet de stocker le plus grand id envoyé par un noeud.
<br></br>
Ensuite à chaque reception d'un message qui n'est pas un message de leader, les noeud compareront l'id recu et leur `maxId`, si ce dernier est plus grand, alors aucun message ne sera envoyé et inversement.

On obtient donc le code suivant:

```go
// définition d'une structure de Noeud
type RingNode struct {
	inChannel  <-chan Message // la connexion entrante
	outChannel chan<- Message // la connexion sortante
	localId    int            // l'id du noeud
	maxId      int            // l'id maximal enregistré par le noeud
	leaderId   int            // l'id du leader
}

func NewRingNode(id int, inChannel <-chan Message, outChannel chan<- Message) *RingNode {
	return &RingNode{inChannel, outChannel, id, id, 0}
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
		} else if receivedMessage.isLeader == false { // si le leader n'as pas encore été idenfié

			// si l'id recu est l'id du noeud courant
			if receivedMessage.id == r.localId {
				// si l'id recu est égal à celui du noeud courant alors on viens d'identifier le leader
				r.leaderId = receivedMessage.id // on met à jour la valeur de l'id du leader
				fmt.Println("From", r.localId, ": Leader found at", r.leaderId)
				r.outChannel <- Message{r.leaderId, true} // on envoie un message contenant l'id du leader et on met à jour la valeur isLeader à true
				close(r.outChannel)                       // on ferme le channel
			} else { // si l'id recu n'est pas l'id courant
				// si l'id recu est plus grand que le plus grand Id envoyé
				if receivedMessage.id > r.maxId {
					max := Max(r.localId, receivedMessage.id)                                                               // on compare les deux id et on renvoie le plus grand des deux
					fmt.Println("From", r.localId, ": compared (", r.localId, "-", receivedMessage.id, ") | sending:", max) // on affiche un message informatif
					r.outChannel <- Message{max, false}
				} // on renvoie le plus grand id des deux
			}

		} else if receivedMessage.isLeader == true { // si le leader est identifié (peut etre remplacé par un else standard)

			r.leaderId = receivedMessage.id // on met à jour la valeur de leaderId avec l'id recu
			fmt.Println("From", r.localId, ": Leader found at", r.leaderId, "| spreading message and closing channel")
			r.outChannel <- receivedMessage // on envoie le message recu au noeud suivant (pas besoin de le modifier)
			close(r.outChannel)             // on ferme la connexion sortante du noeud courant

		}
	}
}
```

Si on veut ajouter un suivi du nombre des messages, il suffit simplement d'intialiser un compteur à 0 juste avant l'appel du main, puis dans la fonction Run, on incrémente ce compteur à chaque message envoyé, puis on affiche la valeur du compteur à la fin du main.

