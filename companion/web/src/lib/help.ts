// The in-app help, one section per screen, in each language. Kept in step
// with docs/PHONE-GUIDE.md (the same guide in English, for people and AIs).
import type { Locale } from './i18n/paraglide/runtime.js';

export type HelpSection = { id: string; title: string; body: string[] };

const en: HelpSection[] = [
	{
		id: 'start',
		title: 'Getting started',
		body: [
			'This app follows a game of The Sims 4 running on a Mac and lets you play it from the phone. The Mac shows a QR code (menu bar icon, then Phone companion). Scan it to open the app; the address changes each time the game starts, so scan again next time.',
			'With VoiceOver, swipe right from the top: first the Pause button, then the tabs, then the screen. Use the rotor on Headings to jump between sections quickly.'
		]
	},
	{
		id: 'top',
		title: 'The top bar',
		body: [
			'The first button pauses the game, or resumes it at the speed it had before. Next to it: the game speed, the day and the time.',
			'Below it are the tabs: Live, Play, House, Sims, History, Settings and Help.',
			'When the game asks something (which recipe, who to invite, a name to type), a card "The game asks" appears right under the screen title on every tab. Choose an answer and press OK, or Cancel.'
		]
	},
	{
		id: 'live',
		title: 'Live',
		body: [
			'"Needs attention" lists what needs doing now: low needs, fire, broken objects, puddles, spoiled food. Each line has buttons that run the game’s own actions, for example Eat now (a quick meal from the fridge) or Cook (opens the recipe question). "Other ways" lists the objects that can help, and the phone.',
			'Then: the Mac voice switch, the household (money, mood of each Sim) and the latest events. "Full history" opens History.'
		]
	},
	{
		id: 'play',
		title: 'Play',
		body: [
			'Interact lists everything the controlled Sim can use: the Sim itself, the Sim’s phone, other Sims, then objects by category (appliances, plumbing, comfort...). Categories are folded: open one with its button. Typing a search opens every category with matches.',
			'Search finds an object by its type or name, in French, Spanish or English ("frigo", "nevera", "fridge"). Press Search on the keyboard to close it and go to the results.',
			'Choose an object to hear its menu, the same as clicking it in the game. Greyed rows say why they are unavailable. Choosing a row adds the action to the Sim’s queue.',
			'Below: which Sim to control, and the four game speeds.'
		]
	},
	{
		id: 'house',
		title: 'House',
		body: [
			'The home floor by floor. Each room says what it is (named from its furniture), where it is on the floor (north, south-east...), its size, and how far it is from the Sim. North is the game’s north.',
			'Open a room to hear the doors and openings to other rooms (tap one to go there), who is inside, and what is in it, with where in the room. Objects open their menu. The garden and the pool are listed too.'
		]
	},
	{
		id: 'sims',
		title: 'Sims',
		body: [
			'Each Sim of the household: job and work performance, mood, needs (lowest first), moodlets, and the action queue, where an action can be cancelled. Skills and Relationships open with their buttons. Relationships go from -100 to 100.'
		]
	},
	{
		id: 'history',
		title: 'History',
		body: [
			'Everything that happened, newest first: moods, moodlets, needs, game messages and questions, money.'
		]
	},
	{
		id: 'settings',
		title: 'Settings',
		body: [
			'Language of the app (game names stay in the game’s language). What VoiceOver announces by itself. "Pause while I choose an action": the game pauses while a menu is open and resumes once the action is queued. The Mac voice switches. Unpair this phone.'
		]
	},
	{
		id: 'mac',
		title: 'On the Mac',
		body: [
			'Option+Command+S turns the Mac’s speech on or off. Option+Command+A opens the TS4 Access menu. The accessibility icon in the macOS menu bar has the same actions and "Phone companion…", which shows the QR code.'
		]
	},
	{
		id: 'trouble',
		title: 'When something goes wrong',
		body: [
			'"Waiting for the game": the game is loading or closed. If the page does not open at all, the game was restarted: scan the new QR code.',
			'An action is not in a menu: the game does not offer it right now (wrong place, Sim busy). Try Refresh in Play.'
		]
	},
	{
		id: 'ai',
		title: 'Ask an AI',
		body: [
			'To talk about this app with ChatGPT or another assistant, ask it to read https://raw.githubusercontent.com/the40n8/Sims4Access-mac/main/llms.txt first.'
		]
	}
];

const fr: HelpSection[] = [
	{
		id: 'start',
		title: 'Pour commencer',
		body: [
			'Cette appli suit une partie des Sims 4 lancée sur un Mac et permet d’y jouer depuis le téléphone. Le Mac affiche un QR code (icône de la barre des menus, puis Phone companion). Scanne-le pour ouvrir l’appli ; l’adresse change à chaque lancement du jeu, il faut donc rescanner la fois suivante.',
			'Avec VoiceOver, balaie vers la droite depuis le haut : d’abord le bouton Pause, puis les onglets, puis l’écran. Le rotor sur Titres permet de sauter d’une section à l’autre.'
		]
	},
	{
		id: 'top',
		title: 'La barre du haut',
		body: [
			'Le premier bouton met le jeu en pause, ou le relance à la vitesse d’avant. À côté : la vitesse, le jour et l’heure du jeu.',
			'En dessous, les onglets : En direct, Jouer, Maison, Sims, Historique, Réglages et Aide.',
			'Quand le jeu pose une question (quelle recette, qui inviter, un nom à taper), une carte « Le jeu demande » apparaît sous le titre, sur tous les onglets. Choisis une réponse puis OK, ou Annuler.'
		]
	},
	{
		id: 'live',
		title: 'En direct',
		body: [
			'« À faire » liste ce qui demande une action : besoins bas, feu, objets cassés, flaques, nourriture avariée. Chaque ligne a des boutons qui lancent les actions du jeu, par exemple Manger (repas rapide au frigo) ou Cuisiner (ouvre la question des recettes). « Autres solutions » liste les objets utiles et le téléphone.',
			'Ensuite : l’interrupteur de la voix du Mac, le foyer (argent, humeur de chaque Sim) et les derniers événements. « Tout l’historique » ouvre l’Historique.'
		]
	},
	{
		id: 'play',
		title: 'Jouer',
		body: [
			'Interagir liste tout ce que le Sim contrôlé peut utiliser : le Sim lui-même, son téléphone, les autres Sims, puis les objets par catégorie (électroménager, plomberie, confort...). Les catégories sont repliées : ouvre-en une avec son bouton. Une recherche ouvre toutes les catégories qui ont des résultats.',
			'La recherche trouve un objet par son type ou son nom, en français, espagnol ou anglais (« frigo », « nevera », « fridge »). Le bouton Rechercher du clavier ferme le clavier et mène aux résultats.',
			'Choisis un objet pour entendre son menu, comme en cliquant dessus dans le jeu. Les lignes grisées disent pourquoi. Choisir une ligne ajoute l’action à la file du Sim.',
			'Plus bas : quel Sim contrôler, et les quatre vitesses du jeu.'
		]
	},
	{
		id: 'house',
		title: 'Maison',
		body: [
			'La maison étage par étage. Chaque pièce dit ce qu’elle est (d’après ses meubles), où elle se trouve (nord, sud-est...), sa taille et sa distance au Sim. Le nord est celui du jeu.',
			'Ouvre une pièce pour entendre les portes et ouvertures vers les autres pièces (touche-en une pour y aller), qui s’y trouve et ce qu’elle contient, avec l’endroit dans la pièce. Les objets ouvrent leur menu. Le jardin et la piscine sont aussi listés.'
		]
	},
	{
		id: 'sims',
		title: 'Sims',
		body: [
			'Chaque Sim du foyer : métier et performance, humeur, besoins (les plus bas d’abord), émotions, et la file d’actions, où l’on peut annuler une action. Compétences et Relations s’ouvrent avec leur bouton. Les relations vont de -100 à 100.'
		]
	},
	{
		id: 'history',
		title: 'Historique',
		body: [
			'Tout ce qui s’est passé, du plus récent au plus ancien : humeurs, émotions, besoins, messages et questions du jeu, argent.'
		]
	},
	{
		id: 'settings',
		title: 'Réglages',
		body: [
			'Langue de l’appli (les noms du jeu restent dans la langue du jeu). Ce que VoiceOver annonce tout seul. « Pause pendant que je choisis une action » : le jeu se met en pause quand un menu est ouvert et repart une fois l’action ajoutée. Les interrupteurs de la voix du Mac. Dissocier ce téléphone.'
		]
	},
	{
		id: 'mac',
		title: 'Sur le Mac',
		body: [
			'Option+Commande+S coupe ou rallume la voix du Mac. Option+Commande+A ouvre le menu TS4 Access. L’icône d’accessibilité dans la barre des menus propose les mêmes actions et « Phone companion… », qui affiche le QR code.'
		]
	},
	{
		id: 'trouble',
		title: 'En cas de souci',
		body: [
			'« En attente du jeu » : le jeu charge ou est fermé. Si la page ne s’ouvre plus du tout, le jeu a été relancé : scanne le nouveau QR code.',
			'Une action manque dans un menu : le jeu ne la propose pas pour l’instant (mauvais endroit, Sim occupé). Essaie Actualiser dans Jouer.'
		]
	},
	{
		id: 'ai',
		title: 'Demander à une IA',
		body: [
			'Pour parler de cette appli avec ChatGPT ou un autre assistant, demande-lui d’abord de lire https://raw.githubusercontent.com/the40n8/Sims4Access-mac/main/llms.txt.'
		]
	}
];

const es: HelpSection[] = [
	{
		id: 'start',
		title: 'Para empezar',
		body: [
			'Esta app sigue una partida de Los Sims 4 en un Mac y permite jugarla desde el teléfono. El Mac muestra un código QR (icono de la barra de menús y luego Phone companion). Escanéalo para abrir la app; la dirección cambia cada vez que se inicia el juego, así que hay que escanear otra vez la próxima vez.',
			'Con VoiceOver, desliza a la derecha desde arriba: primero el botón Pausa, luego las pestañas y luego la pantalla. El rotor en Encabezados permite saltar entre secciones.'
		]
	},
	{
		id: 'top',
		title: 'La barra de arriba',
		body: [
			'El primer botón pausa el juego, o lo reanuda a la velocidad que tenía. Al lado: la velocidad, el día y la hora del juego.',
			'Debajo están las pestañas: En directo, Jugar, Casa, Sims, Historial, Ajustes y Ayuda.',
			'Cuando el juego pregunta algo (qué receta, a quién invitar, un nombre que escribir), aparece una tarjeta «El juego pregunta» bajo el título, en todas las pestañas. Elige una respuesta y pulsa Aceptar, o Cancelar.'
		]
	},
	{
		id: 'live',
		title: 'En directo',
		body: [
			'«Por atender» lista lo que hay que hacer ahora: necesidades bajas, fuego, objetos rotos, charcos, comida podrida. Cada línea tiene botones que lanzan las acciones del juego, por ejemplo Comer (comida rápida de la nevera) o Cocinar (abre la pregunta de recetas). «Otras opciones» lista los objetos que ayudan y el teléfono.',
			'Después: el interruptor de la voz del Mac, el hogar (dinero, ánimo de cada Sim) y los últimos eventos. «Todo el historial» abre el Historial.'
		]
	},
	{
		id: 'play',
		title: 'Jugar',
		body: [
			'Interactuar lista todo lo que el Sim controlado puede usar: el propio Sim, su teléfono, otros Sims y luego los objetos por categoría (electrodomésticos, fontanería, comodidad...). Las categorías están plegadas: abre una con su botón. Una búsqueda abre todas las que tienen resultados.',
			'La búsqueda encuentra un objeto por su tipo o nombre, en español, francés o inglés («nevera», «frigo», «fridge»). El botón Buscar del teclado lo cierra y lleva a los resultados.',
			'Elige un objeto para oír su menú, como al hacer clic en el juego. Las líneas en gris dicen por qué. Elegir una línea añade la acción a la cola del Sim.',
			'Más abajo: qué Sim controlar y las cuatro velocidades del juego.'
		]
	},
	{
		id: 'house',
		title: 'Casa',
		body: [
			'La casa planta por planta. Cada habitación dice qué es (según sus muebles), dónde está (norte, sureste...), su tamaño y su distancia al Sim. El norte es el del juego.',
			'Abre una habitación para oír las puertas y aberturas a otras habitaciones (toca una para ir), quién está dentro y qué contiene, con su lugar en la habitación. Los objetos abren su menú. También aparecen el jardín y la piscina.'
		]
	},
	{
		id: 'sims',
		title: 'Sims',
		body: [
			'Cada Sim del hogar: trabajo y rendimiento, ánimo, necesidades (las más bajas primero), emociones y la cola de acciones, donde se puede cancelar una acción. Habilidades y Relaciones se abren con su botón. Las relaciones van de -100 a 100.'
		]
	},
	{
		id: 'history',
		title: 'Historial',
		body: [
			'Todo lo que ha pasado, de lo más reciente a lo más antiguo: ánimos, emociones, necesidades, mensajes y preguntas del juego, dinero.'
		]
	},
	{
		id: 'settings',
		title: 'Ajustes',
		body: [
			'Idioma de la app (los nombres del juego siguen en el idioma del juego). Lo que VoiceOver anuncia solo. «Pausar mientras elijo una acción»: el juego se pausa mientras hay un menú abierto y sigue cuando la acción se añade. Los interruptores de la voz del Mac. Desvincular este teléfono.'
		]
	},
	{
		id: 'mac',
		title: 'En el Mac',
		body: [
			'Opción+Comando+S apaga o enciende la voz del Mac. Opción+Comando+A abre el menú de TS4 Access. El icono de accesibilidad en la barra de menús tiene las mismas acciones y «Phone companion…», que muestra el código QR.'
		]
	},
	{
		id: 'trouble',
		title: 'Si algo falla',
		body: [
			'«Esperando al juego»: el juego está cargando o cerrado. Si la página ya no abre, el juego se reinició: escanea el nuevo código QR.',
			'Falta una acción en un menú: el juego no la ofrece ahora (lugar equivocado, Sim ocupado). Prueba Actualizar en Jugar.'
		]
	},
	{
		id: 'ai',
		title: 'Preguntar a una IA',
		body: [
			'Para hablar de esta app con ChatGPT u otro asistente, pídele primero que lea https://raw.githubusercontent.com/the40n8/Sims4Access-mac/main/llms.txt.'
		]
	}
];

export const HELP: Record<Locale, HelpSection[]> = { en, fr, es } as Record<Locale, HelpSection[]>;
