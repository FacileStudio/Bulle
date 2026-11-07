import type { Row } from '../row';

export const memory = {
	'account.tab.memory': ['Memory', 'Mémoire', 'Memoria', 'Gedächtnis'],
	'memory.title': ['Memory', 'Mémoire', 'Memoria', 'Gedächtnis'],
	'memory.subtitle': [
		'What the assistant remembers about you across conversations. Read, correct or delete any page.',
		'Ce que l’assistant retient de vous d’une conversation à l’autre. Lisez, corrigez ou supprimez chaque page.',
		'Lo que el asistente recuerda de ti entre conversaciones. Lee, corrige o elimina cualquier página.',
		'Was der Assistent über Gespräche hinweg über Sie weiß. Jede Seite lässt sich lesen, korrigieren oder löschen.'
	],
	'memory.subtitleSpace': [
		'What the assistant remembers in the {space} space. Read, correct or delete any page.',
		'Ce que l’assistant retient dans l’espace {space}. Lisez, corrigez ou supprimez chaque page.',
		'Lo que el asistente recuerda en el espacio {space}. Lee, corrige o elimina cualquier página.',
		'Was der Assistent im Bereich {space} weiß. Jede Seite lässt sich lesen, korrigieren oder löschen.'
	],
	'memory.profile': ['About you', 'À propos de vous', 'Sobre ti', 'Über Sie'],
	'memory.loading': ['Loading memory', 'Chargement de la mémoire', 'Cargando la memoria', 'Gedächtnis wird geladen'],
	'memory.empty.title': ['Nothing remembered yet', 'Rien de retenu pour l’instant', 'Aún no recuerda nada', 'Noch nichts gemerkt'],
	'memory.empty.description': [
		'Tell the assistant about yourself in a chat and it keeps what lasts here.',
		'Parlez de vous à l’assistant dans une conversation : il garde ici ce qui dure.',
		'Háblale de ti al asistente en un chat y guardará aquí lo que perdura.',
		'Erzählen Sie dem Assistenten im Chat von sich, und er behält hier, was bleibt.'
	],
	'memory.openAria': ['Open {page}', 'Ouvrir {page}', 'Abrir {page}', '{page} öffnen'],
	'memory.edit': ['Edit', 'Modifier', 'Editar', 'Bearbeiten'],
	'memory.contentLabel': ['Page content in markdown', 'Contenu de la page en markdown', 'Contenido de la página en markdown', 'Seiteninhalt in Markdown'],
	'memory.saved': ['Memory page saved.', 'Page de mémoire enregistrée.', 'Página de memoria guardada.', 'Gedächtnisseite gespeichert.'],
	'memory.deleted': ['Memory page deleted.', 'Page de mémoire supprimée.', 'Página de memoria eliminada.', 'Gedächtnisseite gelöscht.'],
	'memory.confirm.title': ['Delete {page}?', 'Supprimer {page} ?', '¿Eliminar {page}?', '{page} löschen?'],
	'memory.confirm.description': [
		'The assistant forgets everything on this page. This cannot be undone.',
		'L’assistant oublie tout ce que contient cette page. Action irréversible.',
		'El asistente olvida todo lo que hay en esta página. No se puede deshacer.',
		'Der Assistent vergisst alles auf dieser Seite. Das lässt sich nicht rückgängig machen.'
	],
	'memory.error.connect': [
		'Connect Mycelium in Settings, Apps, to see your memory.',
		'Connectez Mycelium dans Réglages, Applications, pour voir votre mémoire.',
		'Conecta Mycelium en Ajustes, Aplicaciones, para ver tu memoria.',
		'Verbinden Sie Mycelium unter Einstellungen, Apps, um Ihr Gedächtnis zu sehen.'
	],
	'memory.error.missing': [
		'This memory page does not exist any more.',
		'Cette page de mémoire n’existe plus.',
		'Esta página de memoria ya no existe.',
		'Diese Gedächtnisseite existiert nicht mehr.'
	],
	'memory.error.load': ['Could not load your memory.', 'Impossible de charger votre mémoire.', 'No se pudo cargar tu memoria.', 'Gedächtnis konnte nicht geladen werden.'],
	'memory.error.save': ['Could not save the page.', 'Impossible d’enregistrer la page.', 'No se pudo guardar la página.', 'Seite konnte nicht gespeichert werden.'],
	'memory.error.delete': ['Could not delete the page.', 'Impossible de supprimer la page.', 'No se pudo eliminar la página.', 'Seite konnte nicht gelöscht werden.']
} as const satisfies Record<string, Row>;
