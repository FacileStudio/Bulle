import type { Row } from '../row';

export const chat = {
	'chat.new': ['New chat', 'Nouvelle discussion', 'Chat nuevo', 'Neuer Chat'],
	'chat.newTitle': ['New chat - Bulle', 'Nouvelle discussion - Bulle', 'Chat nuevo - Bulle', 'Neuer Chat - Bulle'],
	'chat.defaultTitle': ['Chat', 'Discussion', 'Chat', 'Chat'],
	'chat.conversation': ['Conversation', 'Conversation', 'Conversación', 'Unterhaltung'],
	'chat.emptyTitle': ['What should the agent do?', 'Que doit faire l’agent ?', '¿Qué debe hacer el agente?', 'Was soll der Agent tun?'],
	'chat.emptyBody': [
		'Ask a question or hand over a task. It can use its own computer to get it done.',
		'Posez une question ou confiez une tâche. Il peut utiliser son propre ordinateur pour l’accomplir.',
		'Haz una pregunta o encarga una tarea. Puede usar su propio ordenador para hacerla.',
		'Stellen Sie eine Frage oder übergeben Sie eine Aufgabe. Der Agent kann dafür seinen eigenen Computer nutzen.'
	],
	'chat.taskPlaceholder': ['Describe a task', 'Décrivez une tâche', 'Describe una tarea', 'Beschreiben Sie eine Aufgabe'],
	'chat.startFailed': [
		'Could not start the chat',
		'Impossible de démarrer la discussion',
		'No se pudo iniciar el chat',
		'Der Chat konnte nicht gestartet werden'
	],
	'chat.loadFailed': [
		'Could not load this session',
		'Impossible de charger cette session',
		'No se pudo cargar esta sesión',
		'Diese Sitzung konnte nicht geladen werden'
	],
	'chat.deleted': [
		'This conversation was deleted.',
		'Cette conversation a été supprimée.',
		'Esta conversación se eliminó.',
		'Diese Unterhaltung wurde gelöscht.'
	],
	'chat.nothingTitle': ['Nothing here yet', 'Rien ici pour l’instant', 'Aún no hay nada aquí', 'Noch nichts hier'],
	'chat.nothingBody': [
		'Send a message to start the conversation.',
		'Envoyez un message pour démarrer la conversation.',
		'Envía un mensaje para empezar la conversación.',
		'Senden Sie eine Nachricht, um die Unterhaltung zu beginnen.'
	],
	'chat.errorTitle': ['Something went wrong', 'Une erreur est survenue', 'Algo salió mal', 'Etwas ist schiefgelaufen'],
	'chat.reconnecting': [
		'Reconnecting to the live stream',
		'Reconnexion au flux en direct',
		'Reconectando con la transmisión en directo',
		'Verbindung zum Live-Stream wird wiederhergestellt'
	],
	'chat.connectionLost': [
		'Connection lost. Still trying, and the chat catches up once it is back.',
		'Connexion perdue. Nouvel essai en cours, la discussion se met à jour dès son retour.',
		'Conexión perdida. Seguimos intentándolo y el chat se pondrá al día al volver.',
		'Verbindung verloren. Wird weiter versucht, der Chat holt beim Wiederverbinden auf.'
	],
	'chat.announce.finished': ['Reply finished', 'Réponse terminée', 'Respuesta terminada', 'Antwort fertig'],
	'chat.announce.approval': [
		'Needs your approval',
		'Votre approbation est requise',
		'Necesita tu aprobación',
		'Ihre Freigabe wird benötigt'
	],
	'ask.label': [
		'Questions from the assistant',
		'Questions de l’assistant',
		'Preguntas del asistente',
		'Fragen des Assistenten'
	],
	'ask.pickAny': [
		'Pick any that apply',
		'Choisissez toutes les réponses qui conviennent',
		'Elige todas las que correspondan',
		'Wählen Sie alle zutreffenden'
	],
	'ask.typePlaceholder': ['Type your answer', 'Saisissez votre réponse', 'Escribe tu respuesta', 'Antwort eingeben'],
	'ask.typeLabel': [
		'Type your own answer to: {question}',
		'Saisissez votre propre réponse à : {question}',
		'Escribe tu propia respuesta a: {question}',
		'Eigene Antwort eingeben zu: {question}'
	],
	'ask.answered': ['Answered', 'Répondu', 'Respondida', 'Beantwortet'],
	'ask.send': [
		{ one: 'Send answer', other: 'Send answers' },
		{ one: 'Envoyer la réponse', other: 'Envoyer les réponses' },
		{ one: 'Enviar respuesta', other: 'Enviar respuestas' },
		{ one: 'Antwort senden', other: 'Antworten senden' }
	],
	'chat.sendFailed': [
		'Could not send the message',
		'Impossible d’envoyer le message',
		'No se pudo enviar el mensaje',
		'Die Nachricht konnte nicht gesendet werden'
	],
	'chat.stopFailed': [
		'Could not stop the run',
		'Impossible d’arrêter l’exécution',
		'No se pudo detener la ejecución',
		'Der Lauf konnte nicht gestoppt werden'
	],
	'chat.decideFailed': [
		'Could not send your decision',
		'Impossible d’envoyer votre décision',
		'No se pudo enviar tu decisión',
		'Ihre Entscheidung konnte nicht gesendet werden'
	],
	'chat.computerStatusFailed': [
		'Could not read the computer status',
		'Impossible de lire l’état de l’ordinateur',
		'No se pudo leer el estado del ordenador',
		'Der Status des Computers konnte nicht gelesen werden'
	],
	'composer.placeholder': ['Message the agent', 'Écrire à l’agent', 'Escribe al agente', 'Nachricht an den Agent'],
	'composer.label': ['Message composer', 'Zone de saisie du message', 'Redactor de mensajes', 'Nachrichteneingabe'],
	'composer.message': ['Message', 'Message', 'Mensaje', 'Nachricht'],
	'composer.waiting': [
		'Waiting for uploads to finish',
		'En attente de la fin des téléversements',
		'Esperando a que terminen las subidas',
		'Warten auf den Abschluss der Uploads'
	],
	'composer.removeFailed': [
		'Remove the files that failed to send',
		'Retirez les fichiers qui n’ont pas pu être envoyés',
		'Quita los archivos que no se pudieron enviar',
		'Entfernen Sie die Dateien, die nicht gesendet werden konnten'
	],
	'composer.pastedFile': ['pasted file', 'fichier collé', 'archivo pegado', 'eingefügte Datei'],
	'composer.attached': ['Attached files', 'Fichiers joints', 'Archivos adjuntos', 'Angehängte Dateien'],
	'composer.uploading': ['Uploading {name}', 'Téléversement de {name}', 'Subiendo {name}', '{name} wird hochgeladen'],
	'composer.remove': ['Remove {name}', 'Retirer {name}', 'Quitar {name}', '{name} entfernen'],
	'composer.attachFrom': ['Attach from', 'Joindre depuis', 'Adjuntar desde', 'Anhängen aus'],
	'composer.photo': ['Take a photo', 'Prendre une photo', 'Hacer una foto', 'Foto aufnehmen'],
	'composer.library': ['Photo library', 'Photothèque', 'Fototeca', 'Fotomediathek'],
	'composer.file': ['File', 'Fichier', 'Archivo', 'Datei'],
	'composer.attach': ['Attach files', 'Joindre des fichiers', 'Adjuntar archivos', 'Dateien anhängen'],
	'composer.stop': ['Stop the run', 'Arrêter l’exécution', 'Detener la ejecución', 'Lauf stoppen'],
	'composer.send': ['Send message', 'Envoyer le message', 'Enviar mensaje', 'Nachricht senden'],
	'attach.tooLarge': ['Too large, 20 MB max', 'Trop volumineux, 20 Mo maximum', 'Demasiado grande, 20 MB como máximo', 'Zu groß, maximal 20 MB'],
	'attach.unsupported': ['Unsupported file type', 'Type de fichier non pris en charge', 'Tipo de archivo no admitido', 'Dateityp nicht unterstützt'],
	'attach.tooMany': [
		'{count} files max per message',
		'{count} fichiers maximum par message',
		'{count} archivos como máximo por mensaje',
		'Höchstens {count} Dateien pro Nachricht'
	],
	'attach.failed': ['Upload failed', 'Échec du téléversement', 'Falló la subida', 'Upload fehlgeschlagen'],
	'message.showSteps': ['Show steps', 'Afficher les étapes', 'Mostrar los pasos', 'Schritte anzeigen'],
	'message.hideSteps': ['Hide steps', 'Masquer les étapes', 'Ocultar los pasos', 'Schritte ausblenden'],
	'message.typing': ['Typing', 'En train d’écrire', 'Escribiendo', 'Schreibt'],
	'message.working': ['Working on it', 'Travail en cours', 'Trabajando en ello', 'Arbeitet daran'],
	'message.runFailed': ['The run failed', 'L’exécution a échoué', 'La ejecución falló', 'Der Lauf ist fehlgeschlagen'],
	'thinking.active': ['Thinking', 'Réflexion', 'Pensando', 'Denkt nach'],
	'thinking.done': ['Thought', 'Réflexion terminée', 'Pensamiento', 'Gedacht'],
	'approval.needed': ['Needs your approval', 'Attend votre validation', 'Requiere tu aprobación', 'Wartet auf Ihre Freigabe'],
	'approval.neededLabel': [
		'Approval needed: {tool}',
		'Validation requise : {tool}',
		'Aprobación necesaria: {tool}',
		'Freigabe erforderlich: {tool}'
	],
	'approval.settledLabel': ['Approval: {tool}', 'Validation : {tool}', 'Aprobación: {tool}', 'Freigabe: {tool}'],
	'approval.batchLabel': [
		'Approval needed: {count} x {tool}',
		'Validation requise : {count} x {tool}',
		'Aprobación necesaria: {count} x {tool}',
		'Freigabe erforderlich: {count} x {tool}'
	],
	'approval.approved': ['Approved', 'Validé', 'Aprobado', 'Freigegeben'],
	'approval.denied': ['Denied', 'Refusé', 'Denegado', 'Abgelehnt'],
	'approval.approve': ['Approve', 'Valider', 'Aprobar', 'Freigeben'],
	'approval.deny': ['Deny', 'Refuser', 'Denegar', 'Ablehnen'],
	'approval.approveAll': ['Approve all {count}', 'Tout valider ({count})', 'Aprobar los {count}', 'Alle {count} freigeben'],
	'approval.denyAll': ['Deny all', 'Tout refuser', 'Denegar todo', 'Alle ablehnen'],
	'approval.awaiting': ['Calls awaiting approval', 'Appels en attente de validation', 'Llamadas pendientes de aprobación', 'Aufrufe, die auf Freigabe warten'],
	'approval.remember': [
		'Approve all and stop asking for these same targets in this run',
		'Tout valider et ne plus demander pour ces mêmes cibles pendant cette exécution',
		'Aprobar todo y no volver a preguntar por estos mismos destinos en esta ejecución',
		'Alle freigeben und für dieselben Ziele in diesem Lauf nicht mehr fragen'
	],
	'approval.tool.courrier_send': ['Send an email', 'Envoyer un e-mail', 'Enviar un correo', 'E-Mail senden'],
	'approval.tool.courrier_delete_draft': ['Delete a draft', 'Supprimer un brouillon', 'Eliminar un borrador', 'Entwurf löschen'],
	'approval.tool.agenda_delete_event': ['Delete an event', 'Supprimer un événement', 'Eliminar un evento', 'Termin löschen'],
	'approval.tool.opus_create_task': ['Create a task', 'Créer une tâche', 'Crear una tarea', 'Aufgabe erstellen'],
	'approval.tool.opus_update_task': ['Change a task', 'Modifier une tâche', 'Cambiar una tarea', 'Aufgabe ändern'],
	'approval.tool.opus_delete_task': ['Delete a task', 'Supprimer une tâche', 'Eliminar una tarea', 'Aufgabe löschen'],
	'approval.tool.sablier_start': ['Start a timer', 'Démarrer un chronomètre', 'Iniciar un temporizador', 'Timer starten'],
	'approval.tool.sablier_stop': ['Stop the timer', 'Arrêter le chronomètre', 'Detener el temporizador', 'Timer stoppen'],
	'approval.tool.glouton_save_draft': ['Save a prospect draft', 'Enregistrer un brouillon de prospect', 'Guardar un borrador de prospecto', 'Interessenten-Entwurf speichern'],
	'approval.tool.glouton_update_draft': ['Change a prospect draft', 'Modifier un brouillon de prospect', 'Cambiar un borrador de prospecto', 'Interessenten-Entwurf ändern'],
	'approval.tool.glouton_suggest': ['Suggest a prospect action', 'Suggérer une action de prospection', 'Sugerir una acción de prospección', 'Prospektionsaktion vorschlagen'],
	'approval.tool.glouton_create_company': ['Create a company', 'Créer une entreprise', 'Crear una empresa', 'Unternehmen erstellen'],
	'approval.tool.glouton_update_company': ['Change a company', 'Modifier une entreprise', 'Cambiar una empresa', 'Unternehmen ändern'],
	'approval.tool.glouton_create_person': ['Create a contact', 'Créer un contact', 'Crear un contacto', 'Kontakt erstellen'],
	'approval.tool.glouton_update_person': ['Change a contact', 'Modifier un contact', 'Cambiar un contacto', 'Kontakt ändern'],
	'approval.tool.browser_confirm': ['Send a web request', 'Envoyer une requête web', 'Enviar una solicitud web', 'Webanfrage senden'],
	'approval.tool.opus_create_project': ['Create a project', 'Créer un projet', 'Crear un proyecto', 'Projekt erstellen'],
	'approval.tool.opus_update_project': ['Change a project', 'Modifier un projet', 'Cambiar un proyecto', 'Projekt ändern'],
	'approval.tool.opus_delete_project': ['Delete a project', 'Supprimer un projet', 'Eliminar un proyecto', 'Projekt löschen'],
	'approval.tool.nuage_save': ['Save a file', 'Enregistrer un fichier', 'Guardar un archivo', 'Datei speichern'],
	'approval.tool.jobs_create': ['Schedule a job', 'Planifier une tâche', 'Programar una tarea', 'Job planen'],
	'approval.tool.jobs_update': ['Change a scheduled job', 'Modifier une tâche planifiée', 'Cambiar una tarea programada', 'Geplanten Job ändern'],
	'approval.tool.jobs_delete': ['Delete a scheduled job', 'Supprimer une tâche planifiée', 'Eliminar una tarea programada', 'Geplanten Job löschen'],
	'approval.tool.memory_delete': ['Delete a memory page', 'Supprimer une page de mémoire', 'Eliminar una página de memoria', 'Gedächtnisseite löschen'],
	'approval.tool.bash': ['Run a command', 'Exécuter une commande', 'Ejecutar un comando', 'Befehl ausführen'],
	'approval.serverAsks': [
		'{server} asks you to confirm',
		'{server} vous demande de confirmer',
		'{server} te pide que confirmes',
		'{server} bittet Sie um Bestätigung'
	],
	'approval.openLink': ['Open the page', 'Ouvrir la page', 'Abrir la página', 'Seite öffnen'],
	'tools.kind.apps': ['Facile apps', 'Applications Facile', 'Apps de Facile', 'Facile-Apps'],
	'tools.kind.commands': ['Commands', 'Commandes', 'Comandos', 'Befehle'],
	'tools.kind.files': ['Files', 'Fichiers', 'Archivos', 'Dateien'],
	'tools.kind.web': ['Web', 'Web', 'Web', 'Web'],
	'tools.kind.other': ['Other tools', 'Autres outils', 'Otras herramientas', 'Weitere Werkzeuge'],
	'tools.failed': ['{count} failed', '{count} en échec', '{count} fallidas', '{count} fehlgeschlagen'],
	'tools.running': ['{count} running', '{count} en cours', '{count} en curso', '{count} laufen'],
	'tools.more': ['{count} more', '{count} de plus', '{count} más', '{count} weitere'],
	'tools.phrase.other': [
		{ one: 'used {label}', other: 'used {label} {count} times' },
		{ one: '{label} utilisé', other: '{label} utilisé {count} fois' },
		{ one: 'usó {label}', other: 'usó {label} {count} veces' },
		{ one: '{label} verwendet', other: '{label} {count} Mal verwendet' }
	],
	'tools.phrase.courrier_search_mail': [
		{ one: 'searched mail', other: 'searched mail {count} times' },
		{ one: 'recherche dans les e-mails', other: 'recherche dans les e-mails {count} fois' },
		{ one: 'buscó en el correo', other: 'buscó en el correo {count} veces' },
		{ one: 'E-Mails durchsucht', other: 'E-Mails {count} Mal durchsucht' }
	],
	'tools.phrase.courrier_read_thread': [
		{ one: 'read a conversation', other: 'read {count} conversations' },
		{ one: 'conversation lue', other: '{count} conversations lues' },
		{ one: 'leyó una conversación', other: 'leyó {count} conversaciones' },
		{ one: 'eine Unterhaltung gelesen', other: '{count} Unterhaltungen gelesen' }
	],
	'tools.phrase.courrier_list_accounts': [
		{ one: 'listed mailboxes', other: 'listed mailboxes' },
		{ one: 'boîtes mail listées', other: 'boîtes mail listées' },
		{ one: 'listó los buzones', other: 'listó los buzones' },
		{ one: 'Postfächer aufgelistet', other: 'Postfächer aufgelistet' }
	],
	'tools.phrase.courrier_draft': [
		{ one: 'saved a draft', other: 'saved {count} drafts' },
		{ one: 'brouillon enregistré', other: '{count} brouillons enregistrés' },
		{ one: 'guardó un borrador', other: 'guardó {count} borradores' },
		{ one: 'einen Entwurf gespeichert', other: '{count} Entwürfe gespeichert' }
	],
	'tools.phrase.courrier_send': [
		{ one: 'sent an email', other: 'sent {count} emails' },
		{ one: 'e-mail envoyé', other: '{count} e-mails envoyés' },
		{ one: 'envió un correo', other: 'envió {count} correos' },
		{ one: 'eine E-Mail gesendet', other: '{count} E-Mails gesendet' }
	],
	'tools.phrase.agenda_list_events': [
		{ one: 'checked the calendar', other: 'checked the calendar {count} times' },
		{ one: 'agenda consulté', other: 'agenda consulté {count} fois' },
		{ one: 'revisó el calendario', other: 'revisó el calendario {count} veces' },
		{ one: 'Kalender geprüft', other: 'Kalender {count} Mal geprüft' }
	],
	'tools.phrase.agenda_create_event': [
		{ one: 'created an event', other: 'created {count} events' },
		{ one: 'événement créé', other: '{count} événements créés' },
		{ one: 'creó un evento', other: 'creó {count} eventos' },
		{ one: 'einen Termin erstellt', other: '{count} Termine erstellt' }
	],
	'tools.phrase.opus_list_tasks': [
		{ one: 'checked tasks', other: 'checked tasks {count} times' },
		{ one: 'tâches consultées', other: 'tâches consultées {count} fois' },
		{ one: 'revisó las tareas', other: 'revisó las tareas {count} veces' },
		{ one: 'Aufgaben geprüft', other: 'Aufgaben {count} Mal geprüft' }
	],
	'tools.phrase.opus_create_task': [
		{ one: 'created a task', other: 'created {count} tasks' },
		{ one: 'tâche créée', other: '{count} tâches créées' },
		{ one: 'creó una tarea', other: 'creó {count} tareas' },
		{ one: 'eine Aufgabe erstellt', other: '{count} Aufgaben erstellt' }
	],
	'tools.phrase.glouton_search_contacts': [
		{ one: 'looked up contacts', other: 'looked up contacts {count} times' },
		{ one: 'contacts consultés', other: 'contacts consultés {count} fois' },
		{ one: 'consultó contactos', other: 'consultó contactos {count} veces' },
		{ one: 'Kontakte nachgeschlagen', other: 'Kontakte {count} Mal nachgeschlagen' }
	],
	'tools.phrase.memory_search': [
		{ one: 'searched memory', other: 'searched memory {count} times' },
		{ one: 'recherche dans la mémoire', other: 'recherche dans la mémoire {count} fois' },
		{ one: 'buscó en la memoria', other: 'buscó en la memoria {count} veces' },
		{ one: 'Gedächtnis durchsucht', other: 'Gedächtnis {count} Mal durchsucht' }
	],
	'tools.phrase.memory_write': [
		{ one: 'saved a note', other: 'saved {count} notes' },
		{ one: 'note enregistrée', other: '{count} notes enregistrées' },
		{ one: 'guardó una nota', other: 'guardó {count} notas' },
		{ one: 'eine Notiz gespeichert', other: '{count} Notizen gespeichert' }
	],
	'tools.phrase.memory_delete': [
		{ one: 'forgot a note', other: 'forgot {count} notes' },
		{ one: 'note oubliée', other: '{count} notes oubliées' },
		{ one: 'olvidó una nota', other: 'olvidó {count} notas' },
		{ one: 'eine Notiz vergessen', other: '{count} Notizen vergessen' }
	],
	'tools.phrase.web_search': [
		{ one: 'searched the web', other: 'searched the web {count} times' },
		{ one: 'recherche sur le web', other: 'recherche sur le web {count} fois' },
		{ one: 'buscó en la web', other: 'buscó en la web {count} veces' },
		{ one: 'im Web gesucht', other: '{count} Mal im Web gesucht' }
	],
	'tools.phrase.nuage_save': [
		{ one: 'saved a file', other: 'saved {count} files' },
		{ one: 'fichier enregistré', other: '{count} fichiers enregistrés' },
		{ one: 'guardó un archivo', other: 'guardó {count} archivos' },
		{ one: 'eine Datei gespeichert', other: '{count} Dateien gespeichert' }
	],
	'tools.phrase.nuage_search': [
		{ one: 'searched files', other: 'searched files {count} times' },
		{ one: 'recherche dans les fichiers', other: 'recherche dans les fichiers {count} fois' },
		{ one: 'buscó en los archivos', other: 'buscó en los archivos {count} veces' },
		{ one: 'Dateien durchsucht', other: 'Dateien {count} Mal durchsucht' }
	],
	'tools.phrase.nuage_read': [
		{ one: 'read a file', other: 'read {count} files' },
		{ one: 'fichier lu', other: '{count} fichiers lus' },
		{ one: 'leyó un archivo', other: 'leyó {count} archivos' },
		{ one: 'eine Datei gelesen', other: '{count} Dateien gelesen' }
	],
	'tools.phrase.nuage_list': [
		{ one: 'browsed files', other: 'browsed files {count} times' },
		{ one: 'fichiers parcourus', other: 'fichiers parcourus {count} fois' },
		{ one: 'exploró los archivos', other: 'exploró los archivos {count} veces' },
		{ one: 'Dateien durchgesehen', other: 'Dateien {count} Mal durchgesehen' }
	],
	'tools.phrase.opus_get_task': [
		{ one: 'read a task', other: 'read {count} tasks' },
		{ one: 'tâche lue', other: '{count} tâches lues' },
		{ one: 'leyó una tarea', other: 'leyó {count} tareas' },
		{ one: 'eine Aufgabe gelesen', other: '{count} Aufgaben gelesen' }
	],
	'tools.phrase.opus_update_task': [
		{ one: 'updated a task', other: 'updated {count} tasks' },
		{ one: 'tâche modifiée', other: '{count} tâches modifiées' },
		{ one: 'actualizó una tarea', other: 'actualizó {count} tareas' },
		{ one: 'eine Aufgabe geändert', other: '{count} Aufgaben geändert' }
	],
	'tools.phrase.opus_list_projects': [
		{ one: 'listed projects', other: 'listed projects {count} times' },
		{ one: 'projets listés', other: 'projets listés {count} fois' },
		{ one: 'listó los proyectos', other: 'listó los proyectos {count} veces' },
		{ one: 'Projekte aufgelistet', other: 'Projekte {count} Mal aufgelistet' }
	],
	'tools.phrase.agenda_read_event': [
		{ one: 'read an event', other: 'read {count} events' },
		{ one: 'événement lu', other: '{count} événements lus' },
		{ one: 'leyó un evento', other: 'leyó {count} eventos' },
		{ one: 'einen Termin gelesen', other: '{count} Termine gelesen' }
	],
	'tools.phrase.agenda_find_free_slots': [
		{ one: 'looked for a free slot', other: 'looked for free slots {count} times' },
		{ one: 'créneau libre cherché', other: 'créneaux libres cherchés {count} fois' },
		{ one: 'buscó un hueco libre', other: 'buscó huecos libres {count} veces' },
		{ one: 'freien Termin gesucht', other: '{count} Mal freie Termine gesucht' }
	],
	'tools.phrase.agenda_list_calendars': [
		{ one: 'listed calendars', other: 'listed calendars' },
		{ one: 'agendas listés', other: 'agendas listés' },
		{ one: 'listó los calendarios', other: 'listó los calendarios' },
		{ one: 'Kalender aufgelistet', other: 'Kalender aufgelistet' }
	],
	'tools.phrase.courrier_organize': [
		{ one: 'sorted emails', other: 'sorted emails {count} times' },
		{ one: 'e-mails rangés', other: 'e-mails rangés {count} fois' },
		{ one: 'ordenó correos', other: 'ordenó correos {count} veces' },
		{ one: 'E-Mails sortiert', other: 'E-Mails {count} Mal sortiert' }
	],
	'tools.phrase.glouton_get_company': [
		{ one: 'read a company', other: 'read {count} companies' },
		{ one: 'entreprise consultée', other: '{count} entreprises consultées' },
		{ one: 'consultó una empresa', other: 'consultó {count} empresas' },
		{ one: 'ein Unternehmen gelesen', other: '{count} Unternehmen gelesen' }
	],
	'tools.phrase.glouton_get_person': [
		{ one: 'read a contact', other: 'read {count} contacts' },
		{ one: 'contact consulté', other: '{count} contacts consultés' },
		{ one: 'consultó un contacto', other: 'consultó {count} contactos' },
		{ one: 'einen Kontakt gelesen', other: '{count} Kontakte gelesen' }
	],
	'tools.phrase.memory_read': [
		{ one: 'read a note', other: 'read {count} notes' },
		{ one: 'note lue', other: '{count} notes lues' },
		{ one: 'leyó una nota', other: 'leyó {count} notas' },
		{ one: 'eine Notiz gelesen', other: '{count} Notizen gelesen' }
	],
	'tools.phrase.artifact_create': [
		{ one: 'saved an artifact', other: 'saved {count} artifacts' },
		{ one: 'artefact enregistré', other: '{count} artefacts enregistrés' },
		{ one: 'guardó un artefacto', other: 'guardó {count} artefactos' },
		{ one: 'ein Artefakt gespeichert', other: '{count} Artefakte gespeichert' }
	],
	'tools.phrase.chats_search': [
		{ one: 'searched past chats', other: 'searched past chats {count} times' },
		{ one: 'recherche dans les conversations', other: 'recherche dans les conversations {count} fois' },
		{ one: 'buscó en chats anteriores', other: 'buscó en chats anteriores {count} veces' },
		{ one: 'frühere Chats durchsucht', other: 'frühere Chats {count} Mal durchsucht' }
	],
	'tools.phrase.jobs_create': [
		{ one: 'scheduled a job', other: 'scheduled {count} jobs' },
		{ one: 'tâche planifiée', other: '{count} tâches planifiées' },
		{ one: 'programó una tarea', other: 'programó {count} tareas' },
		{ one: 'einen Job geplant', other: '{count} Jobs geplant' }
	],
	'tools.phrase.jobs_list': [
		{ one: 'listed scheduled jobs', other: 'listed scheduled jobs {count} times' },
		{ one: 'tâches planifiées listées', other: 'tâches planifiées listées {count} fois' },
		{ one: 'listó las tareas programadas', other: 'listó las tareas programadas {count} veces' },
		{ one: 'geplante Jobs aufgelistet', other: 'geplante Jobs {count} Mal aufgelistet' }
	],
	'tools.phrase.bash': [
		{ one: 'ran a command', other: 'ran {count} commands' },
		{ one: 'commande exécutée', other: '{count} commandes exécutées' },
		{ one: 'ejecutó un comando', other: 'ejecutó {count} comandos' },
		{ one: 'einen Befehl ausgeführt', other: '{count} Befehle ausgeführt' }
	],
	'tools.phrase.grep': [
		{ one: 'searched the files', other: 'searched the files {count} times' },
		{ one: 'recherche dans les fichiers', other: 'recherche dans les fichiers {count} fois' },
		{ one: 'buscó en los archivos', other: 'buscó en los archivos {count} veces' },
		{ one: 'Dateien durchsucht', other: 'Dateien {count} Mal durchsucht' }
	],
	'tools.phrase.glob': [
		{ one: 'listed files', other: 'listed files {count} times' },
		{ one: 'fichiers listés', other: 'fichiers listés {count} fois' },
		{ one: 'listó archivos', other: 'listó archivos {count} veces' },
		{ one: 'Dateien aufgelistet', other: 'Dateien {count} Mal aufgelistet' }
	],
	'tools.phrase.todo_write': [
		{ one: 'updated the plan', other: 'updated the plan {count} times' },
		{ one: 'plan mis à jour', other: 'plan mis à jour {count} fois' },
		{ one: 'actualizó el plan', other: 'actualizó el plan {count} veces' },
		{ one: 'Plan aktualisiert', other: 'Plan {count} Mal aktualisiert' }
	]
} as const satisfies Record<string, Row>;

const PREFIX_APPROVAL = 'approval.tool.';
const PREFIX_PHRASE = 'tools.phrase.';

export const APPROVAL_TOOLS = new Set(
	Object.keys(chat)
		.filter((key) => key.startsWith(PREFIX_APPROVAL))
		.map((key) => key.slice(PREFIX_APPROVAL.length))
);

export const PHRASE_TOOLS = new Set(
	Object.keys(chat)
		.filter((key) => key.startsWith(PREFIX_PHRASE) && key !== `${PREFIX_PHRASE}other`)
		.map((key) => key.slice(PREFIX_PHRASE.length))
);
