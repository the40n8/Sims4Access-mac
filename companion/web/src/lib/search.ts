// Search over game names, which are in the game's language (English for us):
// a French or Spanish word also finds the English catalog words it means.

/** Lower case, no accents: "Réfrigérateur" -> "refrigerateur". */
export function fold(text: string): string {
	return text
		.normalize('NFD')
		.replace(/\p{Diacritic}/gu, '')
		.toLocaleLowerCase();
}

// Phone word (folded) -> English catalog words. Prefixes are enough: "frig"
// finds "fridge" whatever comes next.
const WORDS: Record<string, string[]> = {
	// French
	frigo: ['fridge'],
	refrigerateur: ['fridge', 'refrigerator'],
	douche: ['shower'],
	toilette: ['toilet'],
	wc: ['toilet'],
	lit: ['bed'],
	canape: ['sofa', 'couch', 'loveseat'],
	fauteuil: ['chair', 'armchair'],
	chaise: ['chair'],
	tele: ['tv', 'television'],
	television: ['tv', 'television'],
	ordi: ['computer', 'laptop'],
	ordinateur: ['computer', 'laptop'],
	cuisiniere: ['stove', 'oven'],
	four: ['oven', 'stove'],
	plaque: ['stove'],
	evier: ['sink'],
	lavabo: ['sink', 'basin'],
	baignoire: ['bathtub', 'tub'],
	bain: ['bath', 'tub'],
	bureau: ['desk'],
	table: ['table'],
	bibliotheque: ['bookcase', 'bookshelf'],
	livre: ['book'],
	miroir: ['mirror'],
	lampe: ['lamp', 'light'],
	lumiere: ['light', 'lamp'],
	porte: ['door'],
	fenetre: ['window'],
	chevalet: ['easel'],
	echec: ['chess'],
	guitare: ['guitar'],
	piano: ['piano'],
	violon: ['violin'],
	poubelle: ['trash', 'garbage', 'bin'],
	plante: ['plant'],
	jouet: ['toy'],
	berceau: ['crib', 'bassinet'],
	micro: ['microwave', 'microphone'],
	cafe: ['coffee'],
	cafetiere: ['coffee'],
	barbecue: ['barbecue', 'bbq', 'grill'],
	piscine: ['pool'],
	radio: ['stereo', 'radio'],
	chaine: ['stereo'],
	jeu: ['game', 'gaming'],
	telescope: ['telescope', 'observatory'],
	commode: ['dresser'],
	armoire: ['dresser', 'closet', 'wardrobe'],
	placard: ['cabinet', 'cupboard', 'closet'],
	comptoir: ['counter'],
	plan: ['counter'],
	cheminee: ['fireplace'],
	horloge: ['clock'],
	reveil: ['alarm', 'clock'],
	lave: ['dishwasher', 'washer'],
	sport: ['exercise', 'fitness', 'treadmill'],
	velo: ['bike', 'bicycle'],
	tapis: ['rug', 'treadmill'],
	coffre: ['chest', 'toy'],
	boite: ['mailbox', 'box'],
	courrier: ['mailbox'],
	// Spanish
	nevera: ['fridge'],
	frigorifico: ['fridge', 'refrigerator'],
	refrigerador: ['fridge', 'refrigerator'],
	heladera: ['fridge'],
	ducha: ['shower'],
	inodoro: ['toilet'],
	vater: ['toilet'],
	bano: ['toilet', 'bath'],
	cama: ['bed'],
	sofa: ['sofa', 'couch', 'loveseat'],
	sillon: ['chair', 'armchair'],
	silla: ['chair'],
	televisor: ['tv', 'television'],
	ordenador: ['computer', 'laptop'],
	computadora: ['computer', 'laptop'],
	estufa: ['stove', 'oven'],
	horno: ['oven', 'stove'],
	cocina: ['stove', 'oven'],
	fregadero: ['sink'],
	banera: ['bathtub', 'tub'],
	tina: ['bathtub', 'tub'],
	escritorio: ['desk'],
	mesa: ['table'],
	estanteria: ['bookcase', 'bookshelf'],
	librero: ['bookcase', 'bookshelf'],
	libro: ['book'],
	espejo: ['mirror'],
	lampara: ['lamp', 'light'],
	luz: ['light', 'lamp'],
	puerta: ['door'],
	ventana: ['window'],
	caballete: ['easel'],
	ajedrez: ['chess'],
	guitarra: ['guitar'],
	violin: ['violin'],
	basura: ['trash', 'garbage', 'bin'],
	planta: ['plant'],
	juguete: ['toy'],
	cuna: ['crib', 'bassinet'],
	microondas: ['microwave'],
	cafetera: ['coffee'],
	parrilla: ['grill', 'barbecue', 'bbq'],
	piscina: ['pool'],
	juego: ['game', 'gaming'],
	comoda: ['dresser'],
	armario: ['dresser', 'closet', 'wardrobe', 'cabinet'],
	mostrador: ['counter'],
	encimera: ['counter'],
	chimenea: ['fireplace'],
	reloj: ['clock'],
	lavavajillas: ['dishwasher'],
	bici: ['bike', 'bicycle'],
	bicicleta: ['bike', 'bicycle'],
	buzon: ['mailbox']
};

/** Every English word a phone word can mean, plus the word itself. */
export function meanings(word: string): string[] {
	const w = fold(word);
	const out = new Set([w]);
	for (const [key, english] of Object.entries(WORDS)) {
		// "frig" -> frigo; "refrigerateurs" -> refrigerateur.
		if (key.startsWith(w) || (w.startsWith(key) && w.length - key.length <= 2)) {
			for (const e of english) out.add(e);
		}
	}
	return [...out];
}

/** Each typed word must appear (in any of its meanings) in the text. */
export function matchesWords(text: string, query: string): boolean {
	const hay = fold(text);
	const words = query.trim().split(/\s+/).filter(Boolean);
	return words.every((w) => meanings(w).some((m) => hay.includes(m)));
}
