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
	'approval.tool.opus_delete_task': ['Delete a task', 'Supprimer une tâche', 'Eliminar una tarea', 'Aufgabe löschen'],
	'approval.tool.glouton_company_delete': ['Delete a company', 'Supprimer une entreprise', 'Eliminar una empresa', 'Unternehmen löschen'],
	'approval.tool.glouton_person_delete': ['Delete a contact', 'Supprimer un contact', 'Eliminar un contacto', 'Kontakt löschen'],
	'approval.tool.sablier_delete_entry': ['Delete a time entry', 'Supprimer une entrée de temps', 'Eliminar una entrada de tiempo', 'Zeiteintrag löschen'],
	'approval.tool.plume_send_document': ['Send a document for signature', 'Envoyer un document à signer', 'Enviar un documento para firmar', 'Dokument zur Unterschrift senden'],
	'approval.tool.plume_remind_signer': ['Remind a signer', 'Relancer un signataire', 'Recordar a un firmante', 'Unterzeichner erinnern'],
	'approval.tool.plume_delete_document': ['Delete a document to sign', 'Supprimer un document à signer', 'Eliminar un documento para firmar', 'Dokument zum Unterschreiben löschen'],
	'approval.tool.capsule_revoke': ['Destroy a capsule', 'Détruire une capsule', 'Destruir una cápsula', 'Kapsel vernichten'],
	'approval.tool.journal_delete_saved_query': ['Delete a saved query', 'Supprimer une requête enregistrée', 'Eliminar una consulta guardada', 'Gespeicherte Abfrage löschen'],
	'approval.tool.journal_delete_alert': ['Delete an alert rule', "Supprimer une règle d'alerte", 'Eliminar una regla de alerta', 'Alarmregel löschen'],
	'approval.tool.antenne_replay_event': ['Replay an alert event', "Rejouer un événement d'alerte", 'Reenviar un evento de alerta', 'Alarmereignis erneut senden'],
	'approval.tool.vision_delete_site': ['Delete a tracked site', 'Supprimer un site suivi', 'Eliminar un sitio analizado', 'Erfasste Website löschen'],
	'approval.tool.vision_delete_goal': ['Delete a site goal', 'Supprimer un objectif', 'Eliminar un objetivo', 'Website-Ziel löschen'],
	'approval.tool.vision_share_site': ['Open a public dashboard', 'Ouvrir un tableau de bord public', 'Abrir un panel público', 'Öffentliches Dashboard öffnen'],
	'approval.tool.toile_delete_shapes': ['Delete shapes from a board', "Supprimer des formes d'un tableau", 'Eliminar formas de una pizarra', 'Formen von einem Board löschen'],
	'approval.tool.echo_delete_room': ['Delete a meeting room', 'Supprimer une salle de réunion', 'Eliminar una sala de reunión', 'Meetingraum löschen'],
	'approval.tool.browser_confirm': ['Send a web request', 'Envoyer une requête web', 'Enviar una solicitud web', 'Webanfrage senden'],
	'approval.tool.opus_delete_project': ['Delete a project', 'Supprimer un projet', 'Eliminar un proyecto', 'Projekt löschen'],
	'approval.tool.nuage_delete': ['Move to the trash', 'Mettre à la corbeille', 'Mover a la papelera', 'In den Papierkorb verschieben'],
	'approval.tool.nuage_create_share': ['Create a public link', 'Créer un lien public', 'Crear un enlace público', 'Öffentlichen Link erstellen'],
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
	'tools.phrase.opus_search_tasks': [
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
	'tools.phrase.glouton_contacts_search': [
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
	'tools.phrase.sonar_search': [
		{ one: 'searched the web', other: 'searched the web {count} times' },
		{ one: 'recherche sur le web', other: 'recherche sur le web {count} fois' },
		{ one: 'buscó en la web', other: 'buscó en la web {count} veces' },
		{ one: 'im Web gesucht', other: '{count} Mal im Web gesucht' }
	],
	'tools.phrase.sonar_fetch': [
		{ one: 'read a web page', other: 'read {count} web pages' },
		{ one: 'page web lue', other: '{count} pages web lues' },
		{ one: 'leyó una página web', other: 'leyó {count} páginas web' },
		{ one: 'eine Webseite gelesen', other: '{count} Webseiten gelesen' }
	],
	'tools.phrase.nuage_save_file': [
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
	'tools.phrase.nuage_read_file': [
		{ one: 'read a file', other: 'read {count} files' },
		{ one: 'fichier lu', other: '{count} fichiers lus' },
		{ one: 'leyó un archivo', other: 'leyó {count} archivos' },
		{ one: 'eine Datei gelesen', other: '{count} Dateien gelesen' }
	],
	'tools.phrase.nuage_list_folder': [
		{ one: 'browsed files', other: 'browsed files {count} times' },
		{ one: 'fichiers parcourus', other: 'fichiers parcourus {count} fois' },
		{ one: 'exploró los archivos', other: 'exploró los archivos {count} veces' },
		{ one: 'Dateien durchgesehen', other: 'Dateien {count} Mal durchgesehen' }
	],
	'tools.phrase.nuage_create_folder': [
		{ one: 'created a folder', other: 'created {count} folders' },
		{ one: 'dossier créé', other: '{count} dossiers créés' },
		{ one: 'creó una carpeta', other: 'creó {count} carpetas' },
		{ one: 'einen Ordner erstellt', other: '{count} Ordner erstellt' }
	],
	'tools.phrase.nuage_move': [
		{ one: 'moved a file', other: 'moved {count} files' },
		{ one: 'fichier déplacé', other: '{count} fichiers déplacés' },
		{ one: 'movió un archivo', other: 'movió {count} archivos' },
		{ one: 'eine Datei verschoben', other: '{count} Dateien verschoben' }
	],
	'tools.phrase.nuage_delete': [
		{ one: 'moved a file to the trash', other: 'moved {count} files to the trash' },
		{ one: 'fichier mis à la corbeille', other: '{count} fichiers mis à la corbeille' },
		{ one: 'movió un archivo a la papelera', other: 'movió {count} archivos a la papelera' },
		{ one: 'eine Datei in den Papierkorb verschoben', other: '{count} Dateien in den Papierkorb verschoben' }
	],
	'tools.phrase.nuage_create_share': [
		{ one: 'made a public link', other: 'made {count} public links' },
		{ one: 'lien public créé', other: '{count} liens publics créés' },
		{ one: 'creó un enlace público', other: 'creó {count} enlaces públicos' },
		{ one: 'einen öffentlichen Link erstellt', other: '{count} öffentliche Links erstellt' }
	],
	'tools.phrase.nuage_list_shares': [
		{ one: 'listed public links', other: 'listed public links {count} times' },
		{ one: 'liens publics listés', other: 'liens publics listés {count} fois' },
		{ one: 'listó los enlaces públicos', other: 'listó los enlaces públicos {count} veces' },
		{ one: 'öffentliche Links aufgelistet', other: 'öffentliche Links {count} Mal aufgelistet' }
	],
	'tools.phrase.nuage_revoke_share': [
		{ one: 'revoked a public link', other: 'revoked {count} public links' },
		{ one: 'lien public révoqué', other: '{count} liens publics révoqués' },
		{ one: 'revocó un enlace público', other: 'revocó {count} enlaces públicos' },
		{ one: 'einen öffentlichen Link widerrufen', other: '{count} öffentliche Links widerrufen' }
	],
	'tools.phrase.opus_read_task': [
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
	'tools.phrase.sablier_timer_status': [
		{ one: 'checked the timer', other: 'checked the timer {count} times' },
		{ one: 'chronomètre consulté', other: 'chronomètre consulté {count} fois' },
		{ one: 'revisó el temporizador', other: 'revisó el temporizador {count} veces' },
		{ one: 'Timer geprüft', other: 'Timer {count} Mal geprüft' }
	],
	'tools.phrase.sablier_list_entries': [
		{ one: 'read time entries', other: 'read time entries {count} times' },
		{ one: 'entrées de temps lues', other: 'entrées de temps lues {count} fois' },
		{ one: 'leyó las entradas de tiempo', other: 'leyó las entradas de tiempo {count} veces' },
		{ one: 'Zeiteinträge gelesen', other: 'Zeiteinträge {count} Mal gelesen' }
	],
	'tools.phrase.sablier_list_projects': [
		{ one: 'listed time projects', other: 'listed time projects {count} times' },
		{ one: 'projets de temps listés', other: 'projets de temps listés {count} fois' },
		{ one: 'listó los proyectos de tiempo', other: 'listó los proyectos de tiempo {count} veces' },
		{ one: 'Zeitprojekte aufgelistet', other: 'Zeitprojekte {count} Mal aufgelistet' }
	],
	'tools.phrase.sablier_weekly_summary': [
		{ one: 'read the weekly summary', other: 'read {count} weekly summaries' },
		{ one: 'résumé de la semaine lu', other: '{count} résumés de semaine lus' },
		{ one: 'leyó el resumen semanal', other: 'leyó {count} resúmenes semanales' },
		{ one: 'Wochenübersicht gelesen', other: '{count} Wochenübersichten gelesen' }
	],
	'tools.phrase.sablier_start_timer': [
		{ one: 'started the timer', other: 'started the timer {count} times' },
		{ one: 'chronomètre démarré', other: 'chronomètre démarré {count} fois' },
		{ one: 'inició el temporizador', other: 'inició el temporizador {count} veces' },
		{ one: 'Timer gestartet', other: 'Timer {count} Mal gestartet' }
	],
	'tools.phrase.sablier_stop_timer': [
		{ one: 'stopped the timer', other: 'stopped the timer {count} times' },
		{ one: 'chronomètre arrêté', other: 'chronomètre arrêté {count} fois' },
		{ one: 'detuvo el temporizador', other: 'detuvo el temporizador {count} veces' },
		{ one: 'Timer gestoppt', other: 'Timer {count} Mal gestoppt' }
	],
	'tools.phrase.sablier_add_entry': [
		{ one: 'added a time entry', other: 'added {count} time entries' },
		{ one: 'entrée de temps ajoutée', other: '{count} entrées de temps ajoutées' },
		{ one: 'añadió una entrada de tiempo', other: 'añadió {count} entradas de tiempo' },
		{ one: 'einen Zeiteintrag hinzugefügt', other: '{count} Zeiteinträge hinzugefügt' }
	],
	'tools.phrase.sablier_update_entry': [
		{ one: 'edited a time entry', other: 'edited {count} time entries' },
		{ one: 'entrée de temps modifiée', other: '{count} entrées de temps modifiées' },
		{ one: 'editó una entrada de tiempo', other: 'editó {count} entradas de tiempo' },
		{ one: 'einen Zeiteintrag bearbeitet', other: '{count} Zeiteinträge bearbeitet' }
	],
	'tools.phrase.sablier_delete_entry': [
		{ one: 'deleted a time entry', other: 'deleted {count} time entries' },
		{ one: 'entrée de temps supprimée', other: '{count} entrées de temps supprimées' },
		{ one: 'eliminó una entrada de tiempo', other: 'eliminó {count} entradas de tiempo' },
		{ one: 'einen Zeiteintrag gelöscht', other: '{count} Zeiteinträge gelöscht' }
	],
	'tools.phrase.plume_create_document': [
		{ one: 'created a document to sign', other: 'created {count} documents to sign' },
		{ one: 'document à signer créé', other: '{count} documents à signer créés' },
		{ one: 'creó un documento para firmar', other: 'creó {count} documentos para firmar' },
		{ one: 'ein Dokument zum Unterschreiben erstellt', other: '{count} Dokumente zum Unterschreiben erstellt' }
	],
	'tools.phrase.plume_update_document': [
		{ one: 'edited a draft to sign', other: 'edited {count} drafts to sign' },
		{ one: 'brouillon à signer modifié', other: '{count} brouillons à signer modifiés' },
		{ one: 'editó un borrador para firmar', other: 'editó {count} borradores para firmar' },
		{ one: 'einen Entwurf zum Unterschreiben bearbeitet', other: '{count} Entwürfe zum Unterschreiben bearbeitet' }
	],
	'tools.phrase.plume_add_signer': [
		{ one: 'added a signer', other: 'added {count} signers' },
		{ one: 'signataire ajouté', other: '{count} signataires ajoutés' },
		{ one: 'añadió un firmante', other: 'añadió {count} firmantes' },
		{ one: 'einen Unterzeichner hinzugefügt', other: '{count} Unterzeichner hinzugefügt' }
	],
	'tools.phrase.plume_save_client': [
		{ one: 'saved a signing client', other: 'saved {count} signing clients' },
		{ one: 'client de signature enregistré', other: '{count} clients de signature enregistrés' },
		{ one: 'guardó un cliente de firma', other: 'guardó {count} clientes de firma' },
		{ one: 'einen Signaturkunden gespeichert', other: '{count} Signaturkunden gespeichert' }
	],
	'tools.phrase.plume_send_document': [
		{ one: 'sent a document for signature', other: 'sent {count} documents for signature' },
		{ one: 'document envoyé pour signature', other: '{count} documents envoyés pour signature' },
		{ one: 'envió un documento para firmar', other: 'envió {count} documentos para firmar' },
		{ one: 'ein Dokument zur Unterschrift gesendet', other: '{count} Dokumente zur Unterschrift gesendet' }
	],
	'tools.phrase.plume_remind_signer': [
		{ one: 'reminded a signer', other: 'reminded {count} signers' },
		{ one: 'signataire relancé', other: '{count} signataires relancés' },
		{ one: 'recordó a un firmante', other: 'recordó a {count} firmantes' },
		{ one: 'einen Unterzeichner erinnert', other: '{count} Unterzeichner erinnert' }
	],
	'tools.phrase.plume_delete_document': [
		{ one: 'deleted a document to sign', other: 'deleted {count} documents to sign' },
		{ one: 'document à signer supprimé', other: '{count} documents à signer supprimés' },
		{ one: 'eliminó un documento para firmar', other: 'eliminó {count} documentos para firmar' },
		{ one: 'ein Dokument zum Unterschreiben gelöscht', other: '{count} Dokumente zum Unterschreiben gelöscht' }
	],
	'tools.phrase.capsule_revoke': [
		{ one: 'revoked a capsule', other: 'revoked {count} capsules' },
		{ one: 'capsule révoquée', other: '{count} capsules révoquées' },
		{ one: 'revocó una cápsula', other: 'revocó {count} cápsulas' },
		{ one: 'eine Kapsel widerrufen', other: '{count} Kapseln widerrufen' }
	],
	'tools.phrase.journal_save_query': [
		{ one: 'saved a log query', other: 'saved {count} log queries' },
		{ one: 'requête de journal enregistrée', other: '{count} requêtes de journal enregistrées' },
		{ one: 'guardó una consulta de registros', other: 'guardó {count} consultas de registros' },
		{ one: 'eine Log-Abfrage gespeichert', other: '{count} Log-Abfragen gespeichert' }
	],
	'tools.phrase.journal_create_alert': [
		{ one: 'created an alert rule', other: 'created {count} alert rules' },
		{ one: 'règle d\'alerte créée', other: '{count} règles d\'alerte créées' },
		{ one: 'creó una regla de alerta', other: 'creó {count} reglas de alerta' },
		{ one: 'eine Alarmregel erstellt', other: '{count} Alarmregeln erstellt' }
	],
	'tools.phrase.journal_set_alert_enabled': [
		{ one: 'switched an alert rule', other: 'switched {count} alert rules' },
		{ one: 'règle d\'alerte activée ou désactivée', other: '{count} règles d\'alerte activées ou désactivées' },
		{ one: 'activó o desactivó una regla de alerta', other: 'activó o desactivó {count} reglas de alerta' },
		{ one: 'eine Alarmregel umgeschaltet', other: '{count} Alarmregeln umgeschaltet' }
	],
	'tools.phrase.journal_delete_saved_query': [
		{ one: 'deleted a saved query', other: 'deleted {count} saved queries' },
		{ one: 'requête enregistrée supprimée', other: '{count} requêtes enregistrées supprimées' },
		{ one: 'eliminó una consulta guardada', other: 'eliminó {count} consultas guardadas' },
		{ one: 'eine gespeicherte Abfrage gelöscht', other: '{count} gespeicherte Abfragen gelöscht' }
	],
	'tools.phrase.journal_delete_alert': [
		{ one: 'deleted an alert rule', other: 'deleted {count} alert rules' },
		{ one: 'règle d\'alerte supprimée', other: '{count} règles d\'alerte supprimées' },
		{ one: 'eliminó una regla de alerta', other: 'eliminó {count} reglas de alerta' },
		{ one: 'eine Alarmregel gelöscht', other: '{count} Alarmregeln gelöscht' }
	],
	'tools.phrase.antenne_test_target': [
		{ one: 'sent a test alert', other: 'sent {count} test alerts' },
		{ one: 'alerte de test envoyée', other: '{count} alertes de test envoyées' },
		{ one: 'envió una alerta de prueba', other: 'envió {count} alertas de prueba' },
		{ one: 'einen Testalarm gesendet', other: '{count} Testalarme gesendet' }
	],
	'tools.phrase.antenne_set_enabled': [
		{ one: 'switched an alert source or target', other: 'switched {count} alert sources or targets' },
		{ one: 'source ou cible d\'alerte activée ou désactivée', other: '{count} sources ou cibles d\'alerte activées ou désactivées' },
		{ one: 'activó o desactivó una fuente o destino', other: 'activó o desactivó {count} fuentes o destinos' },
		{ one: 'eine Alarmquelle oder ein Ziel umgeschaltet', other: '{count} Alarmquellen oder Ziele umgeschaltet' }
	],
	'tools.phrase.antenne_replay_event': [
		{ one: 'replayed an alert event', other: 'replayed {count} alert events' },
		{ one: 'événement d\'alerte rejoué', other: '{count} événements d\'alerte rejoués' },
		{ one: 'reenvió un evento de alerta', other: 'reenvió {count} eventos de alerta' },
		{ one: 'ein Alarmereignis erneut gesendet', other: '{count} Alarmereignisse erneut gesendet' }
	],
	'tools.phrase.vision_create_site': [
		{ one: 'added a tracked site', other: 'added {count} tracked sites' },
		{ one: 'site suivi ajouté', other: '{count} sites suivis ajoutés' },
		{ one: 'añadió un sitio analizado', other: 'añadió {count} sitios analizados' },
		{ one: 'eine erfasste Website hinzugefügt', other: '{count} erfasste Websites hinzugefügt' }
	],
	'tools.phrase.vision_update_site': [
		{ one: 'edited a tracked site', other: 'edited {count} tracked sites' },
		{ one: 'site suivi modifié', other: '{count} sites suivis modifiés' },
		{ one: 'editó un sitio analizado', other: 'editó {count} sitios analizados' },
		{ one: 'eine erfasste Website bearbeitet', other: '{count} erfasste Websites bearbeitet' }
	],
	'tools.phrase.vision_unshare_site': [
		{ one: 'closed a public dashboard', other: 'closed {count} public dashboards' },
		{ one: 'tableau de bord public fermé', other: '{count} tableaux de bord publics fermés' },
		{ one: 'cerró un panel público', other: 'cerró {count} paneles públicos' },
		{ one: 'ein öffentliches Dashboard geschlossen', other: '{count} öffentliche Dashboards geschlossen' }
	],
	'tools.phrase.vision_share_site': [
		{ one: 'opened a public dashboard', other: 'opened {count} public dashboards' },
		{ one: 'tableau de bord public ouvert', other: '{count} tableaux de bord publics ouverts' },
		{ one: 'abrió un panel público', other: 'abrió {count} paneles públicos' },
		{ one: 'ein öffentliches Dashboard geöffnet', other: '{count} öffentliche Dashboards geöffnet' }
	],
	'tools.phrase.vision_create_goal': [
		{ one: 'added a site goal', other: 'added {count} site goals' },
		{ one: 'objectif ajouté', other: '{count} objectifs ajoutés' },
		{ one: 'añadió un objetivo', other: 'añadió {count} objetivos' },
		{ one: 'ein Website-Ziel hinzugefügt', other: '{count} Website-Ziele hinzugefügt' }
	],
	'tools.phrase.vision_update_goal': [
		{ one: 'edited a site goal', other: 'edited {count} site goals' },
		{ one: 'objectif modifié', other: '{count} objectifs modifiés' },
		{ one: 'editó un objetivo', other: 'editó {count} objetivos' },
		{ one: 'ein Website-Ziel bearbeitet', other: '{count} Website-Ziele bearbeitet' }
	],
	'tools.phrase.vision_delete_site': [
		{ one: 'deleted a tracked site', other: 'deleted {count} tracked sites' },
		{ one: 'site suivi supprimé', other: '{count} sites suivis supprimés' },
		{ one: 'eliminó un sitio analizado', other: 'eliminó {count} sitios analizados' },
		{ one: 'eine erfasste Website gelöscht', other: '{count} erfasste Websites gelöscht' }
	],
	'tools.phrase.vision_delete_goal': [
		{ one: 'deleted a site goal', other: 'deleted {count} site goals' },
		{ one: 'objectif supprimé', other: '{count} objectifs supprimés' },
		{ one: 'eliminó un objetivo', other: 'eliminó {count} objetivos' },
		{ one: 'ein Website-Ziel gelöscht', other: '{count} Website-Ziele gelöscht' }
	],
	'tools.phrase.toile_add_shapes': [
		{ one: 'added shapes to a board', other: 'added shapes to a board {count} times' },
		{ one: 'formes ajoutées au tableau', other: 'formes ajoutées au tableau {count} fois' },
		{ one: 'añadió formas a una pizarra', other: 'añadió formas a una pizarra {count} veces' },
		{ one: 'Formen zu einem Board hinzugefügt', other: '{count} Mal Formen zu einem Board hinzugefügt' }
	],
	'tools.phrase.toile_update_shapes': [
		{ one: 'changed shapes on a board', other: 'changed shapes on a board {count} times' },
		{ one: 'formes du tableau modifiées', other: 'formes du tableau modifiées {count} fois' },
		{ one: 'cambió formas de una pizarra', other: 'cambió formas de una pizarra {count} veces' },
		{ one: 'Formen auf einem Board geändert', other: '{count} Mal Formen auf einem Board geändert' }
	],
	'tools.phrase.toile_delete_shapes': [
		{ one: 'deleted shapes from a board', other: 'deleted shapes from a board {count} times' },
		{ one: 'formes du tableau supprimées', other: 'formes du tableau supprimées {count} fois' },
		{ one: 'eliminó formas de una pizarra', other: 'eliminó formas de una pizarra {count} veces' },
		{ one: 'Formen von einem Board gelöscht', other: '{count} Mal Formen von einem Board gelöscht' }
	],
	'tools.phrase.toile_comment': [
		{ one: 'commented on a board', other: 'left {count} board comments' },
		{ one: 'commentaire ajouté au tableau', other: '{count} commentaires ajoutés au tableau' },
		{ one: 'comentó en una pizarra', other: 'dejó {count} comentarios en pizarras' },
		{ one: 'ein Board kommentiert', other: '{count} Board-Kommentare hinterlassen' }
	],
	'tools.phrase.echo_create_room': [
		{ one: 'created a meeting room', other: 'created {count} meeting rooms' },
		{ one: 'salle de réunion créée', other: '{count} salles de réunion créées' },
		{ one: 'creó una sala de reunión', other: 'creó {count} salas de reunión' },
		{ one: 'einen Meetingraum erstellt', other: '{count} Meetingräume erstellt' }
	],
	'tools.phrase.echo_rename_room': [
		{ one: 'renamed a meeting room', other: 'renamed {count} meeting rooms' },
		{ one: 'salle de réunion renommée', other: '{count} salles de réunion renommées' },
		{ one: 'renombró una sala de reunión', other: 'renombró {count} salas de reunión' },
		{ one: 'einen Meetingraum umbenannt', other: '{count} Meetingräume umbenannt' }
	],
	'tools.phrase.echo_summarize_meeting': [
		{ one: 'summarized a meeting', other: 'summarized {count} meetings' },
		{ one: 'réunion résumée', other: '{count} réunions résumées' },
		{ one: 'resumió una reunión', other: 'resumió {count} reuniones' },
		{ one: 'ein Meeting zusammengefasst', other: '{count} Meetings zusammengefasst' }
	],
	'tools.phrase.echo_share_recording': [
		{ one: 'shared a recording', other: 'shared {count} recordings' },
		{ one: 'enregistrement partagé', other: '{count} enregistrements partagés' },
		{ one: 'compartió una grabación', other: 'compartió {count} grabaciones' },
		{ one: 'eine Aufnahme geteilt', other: '{count} Aufnahmen geteilt' }
	],
	'tools.phrase.echo_delete_room': [
		{ one: 'deleted a meeting room', other: 'deleted {count} meeting rooms' },
		{ one: 'salle de réunion supprimée', other: '{count} salles de réunion supprimées' },
		{ one: 'eliminó una sala de reunión', other: 'eliminó {count} salas de reunión' },
		{ one: 'einen Meetingraum gelöscht', other: '{count} Meetingräume gelöscht' }
	],
	'tools.phrase.plume_list_documents': [
		{ one: 'listed documents to sign', other: 'listed documents to sign {count} times' },
		{ one: 'documents à signer listés', other: 'documents à signer listés {count} fois' },
		{ one: 'listó documentos para firmar', other: 'listó documentos para firmar {count} veces' },
		{ one: 'Dokumente zum Unterschreiben aufgelistet', other: 'Dokumente zum Unterschreiben {count} Mal aufgelistet' }
	],
	'tools.phrase.plume_get_document': [
		{ one: 'checked a document to sign', other: 'checked a document to sign {count} times' },
		{ one: 'document à signer consulté', other: 'document à signer consulté {count} fois' },
		{ one: 'revisó un documento para firmar', other: 'revisó un documento para firmar {count} veces' },
		{ one: 'Dokument zum Unterschreiben geprüft', other: 'Dokument zum Unterschreiben {count} Mal geprüft' }
	],
	'tools.phrase.plume_list_clients': [
		{ one: 'listed signing clients', other: 'listed signing clients {count} times' },
		{ one: 'clients de signature listés', other: 'clients de signature listés {count} fois' },
		{ one: 'listó clientes de firma', other: 'listó clientes de firma {count} veces' },
		{ one: 'Signaturkunden aufgelistet', other: 'Signaturkunden {count} Mal aufgelistet' }
	],
	'tools.phrase.capsule_get_status': [
		{ one: 'checked a capsule', other: 'checked a capsule {count} times' },
		{ one: 'capsule vérifiée', other: 'capsule vérifiée {count} fois' },
		{ one: 'revisó una cápsula', other: 'revisó una cápsula {count} veces' },
		{ one: 'Kapsel geprüft', other: 'Kapsel {count} Mal geprüft' }
	],
	'tools.phrase.journal_search_logs': [
		{ one: 'searched the logs', other: 'searched the logs {count} times' },
		{ one: 'journaux consultés', other: 'journaux consultés {count} fois' },
		{ one: 'buscó en los registros', other: 'buscó en los registros {count} veces' },
		{ one: 'Logs durchsucht', other: 'Logs {count} Mal durchsucht' }
	],
	'tools.phrase.journal_read_log': [
		{ one: 'read a log entry', other: 'read a log entry {count} times' },
		{ one: 'entrée de journal lue', other: 'entrée de journal lue {count} fois' },
		{ one: 'leyó una entrada de registro', other: 'leyó una entrada de registro {count} veces' },
		{ one: 'Logeintrag gelesen', other: 'Logeintrag {count} Mal gelesen' }
	],
	'tools.phrase.journal_count_logs': [
		{ one: 'counted log entries', other: 'counted log entries {count} times' },
		{ one: 'entrées de journal comptées', other: 'entrées de journal comptées {count} fois' },
		{ one: 'contó entradas de registro', other: 'contó entradas de registro {count} veces' },
		{ one: 'Logeinträge gezählt', other: 'Logeinträge {count} Mal gezählt' }
	],
	'tools.phrase.journal_list_apps': [
		{ one: 'listed logged apps', other: 'listed logged apps {count} times' },
		{ one: 'applications journalisées listées', other: 'applications journalisées listées {count} fois' },
		{ one: 'listó las apps registradas', other: 'listó las apps registradas {count} veces' },
		{ one: 'protokollierte Apps aufgelistet', other: 'protokollierte Apps {count} Mal aufgelistet' }
	],
	'tools.phrase.journal_list_saved_queries': [
		{ one: 'listed saved queries', other: 'listed saved queries {count} times' },
		{ one: 'requêtes enregistrées listées', other: 'requêtes enregistrées listées {count} fois' },
		{ one: 'listó consultas guardadas', other: 'listó consultas guardadas {count} veces' },
		{ one: 'gespeicherte Abfragen aufgelistet', other: 'gespeicherte Abfragen {count} Mal aufgelistet' }
	],
	'tools.phrase.journal_list_alerts': [
		{ one: 'listed alert rules', other: 'listed alert rules {count} times' },
		{ one: 'règles d\'alerte listées', other: 'règles d\'alerte listées {count} fois' },
		{ one: 'listó reglas de alerta', other: 'listó reglas de alerta {count} veces' },
		{ one: 'Alarmregeln aufgelistet', other: 'Alarmregeln {count} Mal aufgelistet' }
	],
	'tools.phrase.antenne_list_events': [
		{ one: 'listed alert events', other: 'listed alert events {count} times' },
		{ one: 'événements d\'alerte listés', other: 'événements d\'alerte listés {count} fois' },
		{ one: 'listó eventos de alerta', other: 'listó eventos de alerta {count} veces' },
		{ one: 'Alarmereignisse aufgelistet', other: 'Alarmereignisse {count} Mal aufgelistet' }
	],
	'tools.phrase.antenne_get_event': [
		{ one: 'read an alert event', other: 'read an alert event {count} times' },
		{ one: 'événement d\'alerte lu', other: 'événement d\'alerte lu {count} fois' },
		{ one: 'leyó un evento de alerta', other: 'leyó un evento de alerta {count} veces' },
		{ one: 'Alarmereignis gelesen', other: 'Alarmereignis {count} Mal gelesen' }
	],
	'tools.phrase.antenne_list_providers': [
		{ one: 'listed alert sources', other: 'listed alert sources {count} times' },
		{ one: 'sources d\'alerte listées', other: 'sources d\'alerte listées {count} fois' },
		{ one: 'listó fuentes de alerta', other: 'listó fuentes de alerta {count} veces' },
		{ one: 'Alarmquellen aufgelistet', other: 'Alarmquellen {count} Mal aufgelistet' }
	],
	'tools.phrase.antenne_list_targets': [
		{ one: 'listed delivery targets', other: 'listed delivery targets {count} times' },
		{ one: 'cibles de livraison listées', other: 'cibles de livraison listées {count} fois' },
		{ one: 'listó destinos de entrega', other: 'listó destinos de entrega {count} veces' },
		{ one: 'Zustellziele aufgelistet', other: 'Zustellziele {count} Mal aufgelistet' }
	],
	'tools.phrase.antenne_pool_status': [
		{ one: 'checked the alert bus', other: 'checked the alert bus {count} times' },
		{ one: 'bus d\'alertes vérifié', other: 'bus d\'alertes vérifié {count} fois' },
		{ one: 'revisó el bus de alertas', other: 'revisó el bus de alertas {count} veces' },
		{ one: 'Alarmbus geprüft', other: 'Alarmbus {count} Mal geprüft' }
	],
	'tools.phrase.antenne_test_provider': [
		{ one: 'tested an alert source', other: 'tested an alert source {count} times' },
		{ one: 'source d\'alerte testée', other: 'source d\'alerte testée {count} fois' },
		{ one: 'probó una fuente de alerta', other: 'probó una fuente de alerta {count} veces' },
		{ one: 'Alarmquelle getestet', other: 'Alarmquelle {count} Mal getestet' }
	],
	'tools.phrase.antenne_check_target': [
		{ one: 'checked a delivery target', other: 'checked a delivery target {count} times' },
		{ one: 'cible de livraison vérifiée', other: 'cible de livraison vérifiée {count} fois' },
		{ one: 'revisó un destino de entrega', other: 'revisó un destino de entrega {count} veces' },
		{ one: 'Zustellziel geprüft', other: 'Zustellziel {count} Mal geprüft' }
	],
	'tools.phrase.vision_list_spaces': [
		{ one: 'listed analytics spaces', other: 'listed analytics spaces {count} times' },
		{ one: 'espaces d\'analyse listés', other: 'espaces d\'analyse listés {count} fois' },
		{ one: 'listó espacios de analítica', other: 'listó espacios de analítica {count} veces' },
		{ one: 'Analysebereiche aufgelistet', other: 'Analysebereiche {count} Mal aufgelistet' }
	],
	'tools.phrase.vision_list_sites': [
		{ one: 'listed tracked sites', other: 'listed tracked sites {count} times' },
		{ one: 'sites suivis listés', other: 'sites suivis listés {count} fois' },
		{ one: 'listó sitios analizados', other: 'listó sitios analizados {count} veces' },
		{ one: 'erfasste Websites aufgelistet', other: 'erfasste Websites {count} Mal aufgelistet' }
	],
	'tools.phrase.vision_get_stats': [
		{ one: 'read site analytics', other: 'read site analytics {count} times' },
		{ one: 'statistiques du site lues', other: 'statistiques du site lues {count} fois' },
		{ one: 'leyó la analítica del sitio', other: 'leyó la analítica del sitio {count} veces' },
		{ one: 'Website-Statistiken gelesen', other: 'Website-Statistiken {count} Mal gelesen' }
	],
	'tools.phrase.vision_list_goals': [
		{ one: 'listed site goals', other: 'listed site goals {count} times' },
		{ one: 'objectifs du site listés', other: 'objectifs du site listés {count} fois' },
		{ one: 'listó objetivos del sitio', other: 'listó objetivos del sitio {count} veces' },
		{ one: 'Website-Ziele aufgelistet', other: 'Website-Ziele {count} Mal aufgelistet' }
	],
	'tools.phrase.toile_list_boards': [
		{ one: 'listed boards', other: 'listed boards {count} times' },
		{ one: 'tableaux listés', other: 'tableaux listés {count} fois' },
		{ one: 'listó pizarras', other: 'listó pizarras {count} veces' },
		{ one: 'Boards aufgelistet', other: 'Boards {count} Mal aufgelistet' }
	],
	'tools.phrase.toile_read_board': [
		{ one: 'read a board', other: 'read a board {count} times' },
		{ one: 'tableau lu', other: 'tableau lu {count} fois' },
		{ one: 'leyó una pizarra', other: 'leyó una pizarra {count} veces' },
		{ one: 'Board gelesen', other: 'Board {count} Mal gelesen' }
	],
	'tools.phrase.echo_list_rooms': [
		{ one: 'listed meeting rooms', other: 'listed meeting rooms {count} times' },
		{ one: 'salles de réunion listées', other: 'salles de réunion listées {count} fois' },
		{ one: 'listó salas de reunión', other: 'listó salas de reunión {count} veces' },
		{ one: 'Meetingräume aufgelistet', other: 'Meetingräume {count} Mal aufgelistet' }
	],
	'tools.phrase.echo_list_meetings': [
		{ one: 'listed past meetings', other: 'listed past meetings {count} times' },
		{ one: 'réunions passées listées', other: 'réunions passées listées {count} fois' },
		{ one: 'listó reuniones pasadas', other: 'listó reuniones pasadas {count} veces' },
		{ one: 'vergangene Meetings aufgelistet', other: 'vergangene Meetings {count} Mal aufgelistet' }
	],
	'tools.phrase.echo_get_meeting': [
		{ one: 'checked a meeting', other: 'checked a meeting {count} times' },
		{ one: 'réunion consultée', other: 'réunion consultée {count} fois' },
		{ one: 'revisó una reunión', other: 'revisó una reunión {count} veces' },
		{ one: 'Meeting geprüft', other: 'Meeting {count} Mal geprüft' }
	],
	'tools.phrase.echo_read_transcript': [
		{ one: 'read a transcript', other: 'read a transcript {count} times' },
		{ one: 'transcription lue', other: 'transcription lue {count} fois' },
		{ one: 'leyó una transcripción', other: 'leyó una transcripción {count} veces' },
		{ one: 'Transkript gelesen', other: 'Transkript {count} Mal gelesen' }
	],
	'tools.phrase.opus_list_workspaces': [
		{ one: 'listed workspaces', other: 'listed workspaces {count} times' },
		{ one: 'espaces de travail listés', other: 'espaces de travail listés {count} fois' },
		{ one: 'listó los espacios de trabajo', other: 'listó los espacios de trabajo {count} veces' },
		{ one: 'Arbeitsbereiche aufgelistet', other: 'Arbeitsbereiche {count} Mal aufgelistet' }
	],
	'tools.phrase.opus_comment_task': [
		{ one: 'commented on a task', other: 'commented on {count} tasks' },
		{ one: 'tâche commentée', other: '{count} tâches commentées' },
		{ one: 'comentó una tarea', other: 'comentó {count} tareas' },
		{ one: 'eine Aufgabe kommentiert', other: '{count} Aufgaben kommentiert' }
	],
	'tools.phrase.opus_create_project': [
		{ one: 'created a project', other: 'created {count} projects' },
		{ one: 'projet créé', other: '{count} projets créés' },
		{ one: 'creó un proyecto', other: 'creó {count} proyectos' },
		{ one: 'ein Projekt erstellt', other: '{count} Projekte erstellt' }
	],
	'tools.phrase.opus_update_project': [
		{ one: 'updated a project', other: 'updated {count} projects' },
		{ one: 'projet modifié', other: '{count} projets modifiés' },
		{ one: 'actualizó un proyecto', other: 'actualizó {count} proyectos' },
		{ one: 'ein Projekt geändert', other: '{count} Projekte geändert' }
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
	'tools.phrase.glouton_company_get': [
		{ one: 'read a company', other: 'read {count} companies' },
		{ one: 'entreprise consultée', other: '{count} entreprises consultées' },
		{ one: 'consultó una empresa', other: 'consultó {count} empresas' },
		{ one: 'ein Unternehmen gelesen', other: '{count} Unternehmen gelesen' }
	],
	'tools.phrase.glouton_person_get': [
		{ one: 'read a contact', other: 'read {count} contacts' },
		{ one: 'contact consulté', other: '{count} contacts consultés' },
		{ one: 'consultó un contacto', other: 'consultó {count} contactos' },
		{ one: 'einen Kontakt gelesen', other: '{count} Kontakte gelesen' }
	],
	'tools.phrase.glouton_outreach_list': [
		{ one: 'read past outreach', other: 'read past outreach {count} times' },
		{ one: 'historique de prospection lu', other: 'historique de prospection lu {count} fois' },
		{ one: 'leyó el historial de contacto', other: 'leyó el historial de contacto {count} veces' },
		{ one: 'Kontaktverlauf gelesen', other: 'Kontaktverlauf {count} Mal gelesen' }
	],
	'tools.phrase.glouton_suggestions_list': [
		{ one: 'read prospect suggestions', other: 'read prospect suggestions {count} times' },
		{ one: 'suggestions de prospection lues', other: 'suggestions de prospection lues {count} fois' },
		{ one: 'leyó sugerencias de prospección', other: 'leyó sugerencias de prospección {count} veces' },
		{ one: 'Vorschläge gelesen', other: 'Vorschläge {count} Mal gelesen' }
	],
	'tools.phrase.glouton_draft_save': [
		{ one: 'saved a prospect draft', other: 'saved {count} prospect drafts' },
		{ one: 'brouillon de prospection enregistré', other: '{count} brouillons de prospection enregistrés' },
		{ one: 'guardó un borrador de prospección', other: 'guardó {count} borradores de prospección' },
		{ one: 'einen Entwurf gespeichert', other: '{count} Entwürfe gespeichert' }
	],
	'tools.phrase.glouton_draft_update': [
		{ one: 'changed a prospect draft', other: 'changed {count} prospect drafts' },
		{ one: 'brouillon de prospection modifié', other: '{count} brouillons de prospection modifiés' },
		{ one: 'cambió un borrador de prospección', other: 'cambió {count} borradores de prospección' },
		{ one: 'einen Entwurf geändert', other: '{count} Entwürfe geändert' }
	],
	'tools.phrase.glouton_suggestion_create': [
		{ one: 'suggested a prospect action', other: 'suggested {count} prospect actions' },
		{ one: 'action de prospection suggérée', other: '{count} actions de prospection suggérées' },
		{ one: 'sugirió una acción de prospección', other: 'sugirió {count} acciones de prospección' },
		{ one: 'eine Aktion vorgeschlagen', other: '{count} Aktionen vorgeschlagen' }
	],
	'tools.phrase.glouton_company_create': [
		{ one: 'created a company', other: 'created {count} companies' },
		{ one: 'entreprise créée', other: '{count} entreprises créées' },
		{ one: 'creó una empresa', other: 'creó {count} empresas' },
		{ one: 'ein Unternehmen erstellt', other: '{count} Unternehmen erstellt' }
	],
	'tools.phrase.glouton_company_update': [
		{ one: 'changed a company', other: 'changed {count} companies' },
		{ one: 'entreprise modifiée', other: '{count} entreprises modifiées' },
		{ one: 'cambió una empresa', other: 'cambió {count} empresas' },
		{ one: 'ein Unternehmen geändert', other: '{count} Unternehmen geändert' }
	],
	'tools.phrase.glouton_person_create': [
		{ one: 'created a contact', other: 'created {count} contacts' },
		{ one: 'contact créé', other: '{count} contacts créés' },
		{ one: 'creó un contacto', other: 'creó {count} contactos' },
		{ one: 'einen Kontakt erstellt', other: '{count} Kontakte erstellt' }
	],
	'tools.phrase.glouton_person_update': [
		{ one: 'changed a contact', other: 'changed {count} contacts' },
		{ one: 'contact modifié', other: '{count} contacts modifiés' },
		{ one: 'cambió un contacto', other: 'cambió {count} contactos' },
		{ one: 'einen Kontakt geändert', other: '{count} Kontakte geändert' }
	],
	'tools.phrase.glouton_company_delete': [
		{ one: 'deleted a company', other: 'deleted {count} companies' },
		{ one: 'entreprise supprimée', other: '{count} entreprises supprimées' },
		{ one: 'eliminó una empresa', other: 'eliminó {count} empresas' },
		{ one: 'ein Unternehmen gelöscht', other: '{count} Unternehmen gelöscht' }
	],
	'tools.phrase.glouton_person_delete': [
		{ one: 'deleted a contact', other: 'deleted {count} contacts' },
		{ one: 'contact supprimé', other: '{count} contacts supprimés' },
		{ one: 'eliminó un contacto', other: 'eliminó {count} contactos' },
		{ one: 'einen Kontakt gelöscht', other: '{count} Kontakte gelöscht' }
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
