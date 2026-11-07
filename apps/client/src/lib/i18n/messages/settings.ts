import type { Row } from '../row';

export const settings = {
	'account.pageTitle': ['Settings - Bulle', 'Réglages - Bulle', 'Ajustes - Bulle', 'Einstellungen - Bulle'],
	'account.heading': ['Settings', 'Réglages', 'Ajustes', 'Einstellungen'],
	'account.tabsLabel': ['Settings sections', 'Sections des réglages', 'Secciones de ajustes', 'Einstellungsbereiche'],
	'account.tab.profile': ['Profile', 'Profil', 'Perfil', 'Profil'],
	'account.tab.preferences': ['Preferences', 'Préférences', 'Preferencias', 'Einstellungen'],
	'account.tab.subscription': ['Subscription', 'Abonnement', 'Suscripción', 'Abonnement'],
	'account.plan.free': ['Free plan', 'Offre gratuite', 'Plan gratuito', 'Kostenloser Tarif'],
	'account.plan.paid': ['Paid plan', 'Offre payante', 'Plan de pago', 'Bezahlter Tarif'],
	'account.tab.model': ['Model', 'Modèle', 'Modelo', 'Modell'],
	'account.tab.apps': ['Apps', 'Applications', 'Aplicaciones', 'Apps'],
	'account.tab.jobs': ['Jobs', 'Tâches', 'Tareas', 'Aufgaben'],
	'account.tab.devices': ['Devices', 'Appareils', 'Dispositivos', 'Geräte'],
	'account.tab.notifications': ['Notifications', 'Notifications', 'Notificaciones', 'Benachrichtigungen'],
	'account.tab.updates': ['Updates', 'Mises à jour', 'Actualizaciones', 'Updates'],

	'settings.assistant.title': ['Assistant', 'Assistant', 'Asistente', 'Assistent'],
	'settings.assistant.description': [
		'How your assistant looks, sounds and behaves, here and in the CLI.',
		'Apparence, ton et comportement de votre assistant, ici et dans la CLI.',
		'Aspecto, tono y comportamiento del asistente, aquí y en la CLI.',
		'Aussehen, Ton und Verhalten Ihres Assistenten, hier und in der CLI.'
	],
	'settings.assistant.updated': [
		'Assistant profile updated.',
		'Profil de l’assistant mis à jour.',
		'Perfil del asistente actualizado.',
		'Assistentenprofil aktualisiert.'
	],
	'settings.assistant.saveError': [
		'Could not save assistant profile',
		'Impossible d’enregistrer le profil de l’assistant',
		'No se pudo guardar el perfil del asistente',
		'Assistentenprofil konnte nicht gespeichert werden'
	],
	'settings.assistant.personality': ['Personality', 'Personnalité', 'Personalidad', 'Persönlichkeit'],
	'settings.assistant.traits': ['Traits', 'Traits de caractère', 'Rasgos', 'Eigenschaften'],
	'settings.assistant.save': ['Save Changes', 'Enregistrer les modifications', 'Guardar cambios', 'Änderungen speichern'],

	'computer.title': ['Computer', 'Ordinateur', 'Ordenador', 'Computer'],
	'computer.description': [
		'The private Linux machine where {name} runs code and shell commands.',
		'La machine Linux privée où {name} exécute du code et des commandes shell.',
		'La máquina Linux privada donde {name} ejecuta código y comandos de shell.',
		'Der private Linux-Rechner, auf dem {name} Code und Shell-Befehle ausführt.'
	],
	'computer.status.starting': ['Starting', 'Démarrage', 'Iniciando', 'Startet'],
	'computer.status.startingDesc': [
		'{name} is starting its machine for a task.',
		'{name} démarre sa machine pour une tâche.',
		'{name} está iniciando su máquina para una tarea.',
		'{name} startet seinen Rechner für eine Aufgabe.'
	],
	'computer.status.running': ['Running', 'En marche', 'En ejecución', 'Läuft'],
	'computer.status.runningDesc': [
		'Online. It stops on its own after a while without use.',
		'En ligne. Elle s’arrête seule après un moment d’inactivité.',
		'En línea. Se detiene sola tras un tiempo sin uso.',
		'Online. Er stoppt von selbst, wenn er eine Weile nicht genutzt wird.'
	],
	'computer.status.error': ['Error', 'Erreur', 'Error', 'Fehler'],
	'computer.status.errorDesc': [
		'The machine hit a problem. The next task that needs it retries.',
		'La machine a rencontré un problème. La prochaine tâche qui en a besoin réessaie.',
		'La máquina tuvo un problema. La próxima tarea que la necesite lo reintentará.',
		'Der Rechner hatte ein Problem. Die nächste Aufgabe, die ihn braucht, versucht es erneut.'
	],
	'computer.status.asleep': ['Asleep', 'En veille', 'En reposo', 'Im Ruhezustand'],
	'computer.status.asleepDesc': [
		'{name} starts it on its own when a task needs it.',
		'{name} la démarre seul quand une tâche en a besoin.',
		'{name} la inicia sola cuando una tarea la necesita.',
		'{name} startet ihn von selbst, sobald eine Aufgabe ihn braucht.'
	],
	'computer.stop': ['Stop now', 'Arrêter maintenant', 'Detener ahora', 'Jetzt stoppen'],
	'computer.spec.cpu': ['CPU', 'CPU', 'CPU', 'CPU'],
	'computer.spec.cpuValue': ['{count} vCPU', '{count} vCPU', '{count} vCPU', '{count} vCPU'],
	'computer.spec.memory': ['Memory', 'Mémoire', 'Memoria', 'Arbeitsspeicher'],
	'computer.spec.memoryValue': ['{count} GB', '{count} Go', '{count} GB', '{count} GB'],
	'computer.spec.disk': ['Disk', 'Disque', 'Disco', 'Festplatte'],
	'computer.spec.diskValue': ['{count} GB', '{count} Go', '{count} GB', '{count} GB'],
	'computer.spec.lastActivity': ['Last activity', 'Dernière activité', 'Última actividad', 'Letzte Aktivität'],
	'computer.spec.never': ['Never', 'Jamais', 'Nunca', 'Nie'],
	'computer.network.label': ['Offline', 'Hors ligne', 'Sin conexión', 'Offline'],
	'computer.network.description': [
		'Offline blocks installs and web access from its shell. web_fetch still works, because it runs outside the computer. Changing this restarts the computer if it is on.',
		'Hors ligne bloque les installations et l’accès au web depuis son terminal. web_fetch fonctionne toujours, car il s’exécute hors de l’ordinateur. Changer ce réglage redémarre l’ordinateur s’il est allumé.',
		'Sin conexión bloquea las instalaciones y el acceso web desde su terminal. web_fetch sigue funcionando, porque se ejecuta fuera del ordenador. Cambiarlo reinicia el ordenador si está encendido.',
		'Offline blockiert Installationen und Webzugriff aus seiner Shell. web_fetch funktioniert weiter, weil es außerhalb des Computers läuft. Eine Änderung startet den Computer neu, wenn er läuft.'
	],

	'hardware.title': ['Devices', 'Appareils', 'Dispositivos', 'Geräte'],
	'hardware.description': [
		'Pastille and other BOHS-01 devices paired to your account. Bulle relays a paired device over Bluetooth while this page is open; with Wi-Fi set, it reaches Bulle on its own.',
		'Pastille et autres appareils BOHS-01 associés à votre compte. Bulle relaie un appareil associé en Bluetooth tant que cette page est ouverte ; avec le Wi-Fi configuré, il joint Bulle seul.',
		'Pastille y otros dispositivos BOHS-01 vinculados a la cuenta. Bulle retransmite un dispositivo vinculado por Bluetooth mientras esta página está abierta; con Wi-Fi configurado, se conecta a Bulle por sí solo.',
		'Pastille und andere BOHS-01-Geräte, die mit Ihrem Konto gekoppelt sind. Bulle leitet ein gekoppeltes Gerät per Bluetooth weiter, solange diese Seite offen ist; mit eingerichtetem WLAN erreicht es Bulle selbstständig.'
	],
	'hardware.pair': ['Pair device', 'Associer un appareil', 'Vincular dispositivo', 'Gerät koppeln'],
	'hardware.pairing': ['Pairing…', 'Association…', 'Vinculando…', 'Kopplung…'],
	'hardware.paired': ['{name} paired', '{name} associé', '{name} vinculado', '{name} gekoppelt'],
	'hardware.pairFailed': [
		'Could not pair the device',
		'Impossible d’associer l’appareil',
		'No se pudo vincular el dispositivo',
		'Gerät konnte nicht gekoppelt werden'
	],
	'hardware.connectFailed': [
		'Could not connect the device',
		'Impossible de connecter l’appareil',
		'No se pudo conectar el dispositivo',
		'Gerät konnte nicht verbunden werden'
	],
	'hardware.loadFailed': [
		'Could not load your devices',
		'Impossible de charger vos appareils',
		'No se pudieron cargar los dispositivos',
		'Geräte konnten nicht geladen werden'
	],
	'hardware.revokeFailed': [
		'Could not revoke the device',
		'Impossible de révoquer l’appareil',
		'No se pudo revocar el dispositivo',
		'Gerät konnte nicht widerrufen werden'
	],
	'hardware.revoked': ['{name} revoked', '{name} révoqué', '{name} revocado', '{name} widerrufen'],
	'hardware.unsupportedApp': [
		'Pair and relay devices from Bulle in Chrome or Edge on the web.',
		'Associez et relayez vos appareils depuis Bulle dans Chrome ou Edge sur le web.',
		'Vincula y retransmite dispositivos desde Bulle en Chrome o Edge en la web.',
		'Koppeln und Weiterleiten von Geräten ist mit Bulle im Web in Chrome oder Edge möglich.'
	],
	'hardware.unsupportedBrowser': [
		'This browser has no Web Bluetooth. Open Bulle in Chrome or Edge to pair a device.',
		'Ce navigateur ne gère pas le Web Bluetooth. Ouvrez Bulle dans Chrome ou Edge pour associer un appareil.',
		'Este navegador no tiene Web Bluetooth. Abre Bulle en Chrome o Edge para vincular un dispositivo.',
		'Dieser Browser unterstützt kein Web Bluetooth. Öffnen Sie Bulle in Chrome oder Edge, um ein Gerät zu koppeln.'
	],
	'hardware.remote.title': [
		'Add a remote device',
		'Ajouter un appareil à distance',
		'Añadir un dispositivo remoto',
		'Entferntes Gerät hinzufügen'
	],
	'hardware.remote.help': [
		'For a device with no Bluetooth link to this app, such as a Raspberry Pi at home. Enter the code it shows.',
		'Pour un appareil sans lien Bluetooth avec cette application, comme un Raspberry Pi à la maison. Saisissez le code qu’il affiche.',
		'Para un dispositivo sin enlace Bluetooth con esta app, como una Raspberry Pi en casa. Introduce el código que muestra.',
		'Für ein Gerät ohne Bluetooth-Verbindung zu dieser App, etwa einen Raspberry Pi zu Hause. Geben Sie den angezeigten Code ein.'
	],
	'hardware.remote.code': ['Pairing code', 'Code d’association', 'Código de vinculación', 'Kopplungscode'],
	'hardware.remote.name': [
		'Name (optional)',
		'Nom (facultatif)',
		'Nombre (opcional)',
		'Name (optional)'
	],
	'hardware.remote.nameHelp': [
		'Bulle finds the device by this name, as in "turn on the living room light".',
		'Bulle retrouve l’appareil par ce nom, par exemple « allume la lumière du salon ».',
		'Bulle encuentra el dispositivo por este nombre, como en «enciende la luz del salón».',
		'Bulle findet das Gerät über diesen Namen, etwa „schalte das Wohnzimmerlicht ein“.'
	],
	'hardware.remote.add': ['Add device', 'Ajouter l’appareil', 'Añadir dispositivo', 'Gerät hinzufügen'],
	'hardware.remote.failed': [
		'Could not add the device',
		'Impossible d’ajouter l’appareil',
		'No se pudo añadir el dispositivo',
		'Gerät konnte nicht hinzugefügt werden'
	],
	'hardware.wifi': ['Wi-Fi', 'Wi-Fi', 'Wi-Fi', 'WLAN'],
	'hardware.wifiHelp': [
		'Optional. Both fields go to your Pastille when you pair it, so it reaches Bulle without this page open.',
		'Facultatif. Les deux champs sont transmis à votre Pastille lors de l’association, pour qu’elle joigne Bulle sans cette page ouverte.',
		'Opcional. Ambos campos se envían a la Pastille al vincularla, para que llegue a Bulle sin esta página abierta.',
		'Optional. Beide Felder gehen bei der Kopplung an Ihre Pastille, damit sie Bulle auch ohne diese offene Seite erreicht.'
	],
	'hardware.wifiNetwork': [
		'Wi-Fi network (optional)',
		'Réseau Wi-Fi (facultatif)',
		'Red Wi-Fi (opcional)',
		'WLAN-Netzwerk (optional)'
	],
	'hardware.wifiPassword': ['Wi-Fi password', 'Mot de passe Wi-Fi', 'Contraseña de Wi-Fi', 'WLAN-Passwort'],
	'hardware.seen': ['Seen {time}', 'Vu {time}', 'Visto {time}', 'Zuletzt gesehen {time}'],
	'hardware.neverConnected': ['Never connected', 'Jamais connecté', 'Nunca conectado', 'Nie verbunden'],
	'hardware.revoke': ['Revoke', 'Révoquer', 'Revocar', 'Widerrufen'],
	'hardware.empty.title': [
		'No device paired yet.',
		'Aucun appareil associé pour le moment.',
		'Aún no hay ningún dispositivo vinculado.',
		'Noch kein Gerät gekoppelt.'
	],
	'hardware.empty.hint': [
		'Pair a Pastille over Bluetooth to get started.',
		'Associez une Pastille en Bluetooth pour commencer.',
		'Vincula una Pastille por Bluetooth para empezar.',
		'Koppeln Sie eine Pastille per Bluetooth, um zu starten.'
	],
	'hardware.revokeTitle': ['Revoke {name}?', 'Révoquer {name} ?', '¿Revocar {name}?', '{name} widerrufen?'],
	'hardware.revokeTitleFallback': [
		'Revoke this device?',
		'Révoquer cet appareil ?',
		'¿Revocar este dispositivo?',
		'Dieses Gerät widerrufen?'
	],
	'hardware.revokeDescription': [
		'Its token stops working at once. Pair it again to use it.',
		'Son jeton cesse de fonctionner immédiatement. Associez-le de nouveau pour l’utiliser.',
		'Su token deja de funcionar de inmediato. Vincúlalo de nuevo para usarlo.',
		'Das Token funktioniert sofort nicht mehr. Koppeln Sie das Gerät erneut, um es zu nutzen.'
	],
	'hardware.revokeConfirm': ['Revoke device', 'Révoquer l’appareil', 'Revocar dispositivo', 'Gerät widerrufen'],

	'jobs.title': ['Jobs', 'Tâches', 'Tareas', 'Aufgaben'],
	'jobs.subtitleSpace': [
		'Tasks your assistant runs on a schedule in {space}.',
		'Tâches que votre assistant exécute selon un calendrier dans {space}.',
		'Tareas que el asistente ejecuta de forma programada en {space}.',
		'Aufgaben, die Ihr Assistent in {space} nach Zeitplan ausführt.'
	],
	'jobs.subtitle': [
		'Tasks your assistant runs on a schedule.',
		'Tâches que votre assistant exécute selon un calendrier.',
		'Tareas que el asistente ejecuta de forma programada.',
		'Aufgaben, die Ihr Assistent nach Zeitplan ausführt.'
	],
	'jobs.new': ['New job', 'Nouvelle tâche', 'Nueva tarea', 'Neue Aufgabe'],
	'jobs.loadFailed': [
		'Failed to load background jobs',
		'Échec du chargement des tâches de fond',
		'Error al cargar las tareas en segundo plano',
		'Hintergrundaufgaben konnten nicht geladen werden'
	],
	'jobs.empty.title': [
		'No scheduled jobs',
		'Aucune tâche planifiée',
		'No hay tareas programadas',
		'Keine geplanten Aufgaben'
	],
	'jobs.empty.description': [
		'Set up scheduled background jobs to perform regular analyses, summaries, or reports.',
		'Planifiez des tâches de fond pour produire régulièrement des analyses, des synthèses ou des rapports.',
		'Configura tareas programadas en segundo plano para generar análisis, resúmenes o informes periódicos.',
		'Richten Sie geplante Hintergrundaufgaben für regelmäßige Analysen, Zusammenfassungen oder Berichte ein.'
	],
	'jobs.empty.create': [
		'Create your first job',
		'Créer votre première tâche',
		'Crear la primera tarea',
		'Erste Aufgabe erstellen'
	],
	'jobs.running': ['Running', 'En cours', 'En ejecución', 'Läuft'],
	'jobs.toggleAria': ['Toggle {name}', 'Activer ou désactiver {name}', 'Activar o desactivar {name}', '{name} umschalten'],
	'jobs.runNow': ['Run now', 'Lancer maintenant', 'Ejecutar ahora', 'Jetzt ausführen'],
	'jobs.watch': ['Watch', 'Suivre', 'Seguir', 'Verfolgen'],
	'jobs.viewOutput': ['View output', 'Voir le résultat', 'Ver resultado', 'Ergebnis ansehen'],
	'jobs.runsOf': ['Runs of {name}', 'Exécutions de {name}', 'Ejecuciones de {name}', 'Ausführungen von {name}'],
	'jobs.editAria': ['Edit {name}', 'Modifier {name}', 'Editar {name}', '{name} bearbeiten'],
	'jobs.deleteAria': ['Delete {name}', 'Supprimer {name}', 'Eliminar {name}', '{name} löschen'],
	'jobs.createTitle': [
		'Create Background Job',
		'Créer une tâche de fond',
		'Crear tarea en segundo plano',
		'Hintergrundaufgabe erstellen'
	],
	'jobs.editTitle': ['Edit Job: {name}', 'Modifier la tâche : {name}', 'Editar tarea: {name}', 'Aufgabe bearbeiten: {name}'],
	'jobs.field.name': ['Name', 'Nom', 'Nombre', 'Name'],
	'jobs.field.namePlaceholder': ['e.g. daily-summary', 'ex. synthese-quotidienne', 'p. ej. resumen-diario', 'z. B. tagesuebersicht'],
	'jobs.field.schedule': [
		'Schedule (cron expression)',
		'Planification (expression cron)',
		'Programación (expresión cron)',
		'Zeitplan (Cron-Ausdruck)'
	],
	'jobs.field.scheduleHelp': [
		'Standard 5-part cron or macros like @hourly, @daily',
		'Cron standard à 5 champs ou macros comme @hourly, @daily',
		'Cron estándar de 5 campos o macros como @hourly, @daily',
		'Standard-Cron mit 5 Feldern oder Makros wie @hourly, @daily'
	],
	'jobs.field.prompt': ['Prompt', 'Consigne', 'Instrucción', 'Anweisung'],
	'jobs.field.promptHelp': [
		'The instruction given to Bulle on every scheduled execution',
		'L’instruction donnée à Bulle à chaque exécution planifiée',
		'La instrucción que recibe Bulle en cada ejecución programada',
		'Die Anweisung, die Bulle bei jeder geplanten Ausführung erhält'
	],
	'jobs.field.promptPlaceholder': [
		'e.g. Check for urgent incoming messages and summarize top priorities',
		'ex. Vérifier les messages urgents reçus et résumer les priorités',
		'p. ej. Revisar los mensajes urgentes entrantes y resumir las prioridades',
		'z. B. Dringende eingehende Nachrichten prüfen und die wichtigsten Prioritäten zusammenfassen'
	],
	'jobs.field.profile': ['Profile (optional)', 'Profil (facultatif)', 'Perfil (opcional)', 'Profil (optional)'],
	'jobs.field.delivery': ['Delivery', 'Livraison', 'Entrega', 'Zustellung'],
	'jobs.delivery.none': ['None', 'Aucune', 'Ninguna', 'Keine'],
	'jobs.delivery.push': ['Push notification', 'Notification push', 'Notificación push', 'Push-Benachrichtigung'],
	'jobs.delivery.legacy': ['{value} (legacy)', '{value} (ancien)', '{value} (heredado)', '{value} (veraltet)'],
	'jobs.enableNow': [
		'Enable job immediately',
		'Activer la tâche immédiatement',
		'Activar la tarea de inmediato',
		'Aufgabe sofort aktivieren'
	],
	'jobs.enableAria': ['Enable job', 'Activer la tâche', 'Activar tarea', 'Aufgabe aktivieren'],
	'jobs.active': ['Active', 'Active', 'Activa', 'Aktiv'],
	'jobs.creating': ['Creating...', 'Création…', 'Creando…', 'Wird erstellt…'],
	'jobs.createSubmit': ['Create Job', 'Créer la tâche', 'Crear tarea', 'Aufgabe erstellen'],
	'jobs.saving': ['Saving...', 'Enregistrement…', 'Guardando…', 'Wird gespeichert…'],
	'jobs.saveChanges': ['Save Changes', 'Enregistrer les modifications', 'Guardar cambios', 'Änderungen speichern'],
	'jobs.created': ['Job created', 'Tâche créée', 'Tarea creada', 'Aufgabe erstellt'],
	'jobs.createFailed': [
		'Failed to create job',
		'Échec de la création de la tâche',
		'Error al crear la tarea',
		'Aufgabe konnte nicht erstellt werden'
	],
	'jobs.updated': ['Job updated', 'Tâche mise à jour', 'Tarea actualizada', 'Aufgabe aktualisiert'],
	'jobs.updateFailed': [
		'Failed to update job',
		'Échec de la mise à jour de la tâche',
		'Error al actualizar la tarea',
		'Aufgabe konnte nicht aktualisiert werden'
	],
	'jobs.enabled': ['Job enabled', 'Tâche activée', 'Tarea activada', 'Aufgabe aktiviert'],
	'jobs.disabled': ['Job disabled', 'Tâche désactivée', 'Tarea desactivada', 'Aufgabe deaktiviert'],
	'jobs.toggleFailed': [
		'Failed to toggle job state',
		'Échec du changement d’état de la tâche',
		'Error al cambiar el estado de la tarea',
		'Aufgabenstatus konnte nicht geändert werden'
	],
	'jobs.deleteConfirm': [
		'Delete background job "{name}"?',
		'Supprimer la tâche de fond « {name} » ?',
		'¿Eliminar la tarea en segundo plano «{name}»?',
		'Hintergrundaufgabe „{name}“ löschen?'
	],
	'jobs.deleted': ['Job deleted', 'Tâche supprimée', 'Tarea eliminada', 'Aufgabe gelöscht'],
	'jobs.deleteFailed': [
		'Failed to delete job',
		'Échec de la suppression de la tâche',
		'Error al eliminar la tarea',
		'Aufgabe konnte nicht gelöscht werden'
	],
	'jobs.runFailed': [
		'Failed to start job run',
		'Échec du lancement de la tâche',
		'Error al iniciar la ejecución de la tarea',
		'Aufgabenausführung konnte nicht gestartet werden'
	],
	'jobs.historyFailed': [
		'Failed to load run history',
		'Échec du chargement de l’historique',
		'Error al cargar el historial de ejecuciones',
		'Ausführungsverlauf konnte nicht geladen werden'
	],
	'jobs.historyTitle': ['Runs', 'Exécutions', 'Ejecuciones', 'Ausführungen'],
	'jobs.history.empty': ['No runs recorded', 'Aucune exécution enregistrée', 'No hay ejecuciones registradas', 'Keine Ausführungen aufgezeichnet'],
	'jobs.history.emptyDescription': [
		"This job has not been executed yet. Use 'Run now' to trigger an immediate test run.",
		'Cette tâche n’a pas encore été exécutée. Utilisez « Lancer maintenant » pour déclencher un essai immédiat.',
		'Esta tarea aún no se ha ejecutado. Usa «Ejecutar ahora» para lanzar una prueba inmediata.',
		'Diese Aufgabe wurde noch nicht ausgeführt. Mit „Jetzt ausführen“ starten Sie sofort einen Testlauf.'
	],
	'jobs.history.inProgress': ['In progress', 'En cours', 'En curso', 'In Arbeit'],
	'jobs.status.finished': ['Finished', 'Terminée', 'Finalizada', 'Abgeschlossen'],
	'jobs.status.running': ['Running', 'En cours', 'En ejecución', 'Läuft'],
	'jobs.status.error': ['Error', 'Erreur', 'Error', 'Fehler'],
	'jobs.schedule.everyDay': ['Every day', 'Tous les jours', 'Todos los días', 'Jeden Tag'],
	'jobs.schedule.weekdays': ['Weekdays', 'En semaine', 'Días laborables', 'Werktags'],
	'jobs.schedule.weekends': ['Weekends', 'Le week-end', 'Fines de semana', 'Am Wochenende'],
	'jobs.schedule.at': ['{days} at {time}', '{days} à {time}', '{days} a las {time}', '{days} um {time}'],

	'matrix.title': ['Chat agent', 'Agent de discussion', 'Agente de chat', 'Chat-Agent'],
	'matrix.description': [
		'Your assistant is a member of every channel of your Prisme spaces. Mention it in a channel and it answers in a thread, working in the matching space. There is nothing to set up.',
		'Votre assistant est membre de chaque canal de vos espaces Prisme. Mentionnez-le dans un canal : il répond dans un fil, en travaillant dans l’espace correspondant. Rien à configurer.',
		'El asistente es miembro de cada canal de tus espacios de Prisme. Menciónalo en un canal y responde en un hilo, trabajando en el espacio correspondiente. No hay nada que configurar.',
		'Ihr Assistent ist Mitglied jedes Kanals Ihrer Prisme-Bereiche. Erwähnen Sie ihn in einem Kanal, und er antwortet in einem Thread und arbeitet im passenden Bereich. Es ist nichts einzurichten.'
	],
	'matrix.loadFailed': [
		'Could not load the chat agent',
		'Impossible de charger l’agent de discussion',
		'No se pudo cargar el agente de chat',
		'Chat-Agent konnte nicht geladen werden'
	],
	'matrix.defaultSpace': ['Default space', 'Espace par défaut', 'Espacio predeterminado', 'Standardbereich'],
	'matrix.unavailable': [
		'Chat agents are not enabled on this Bulle server.',
		'Les agents de discussion ne sont pas activés sur ce serveur Bulle.',
		'Los agentes de chat no están activados en este servidor de Bulle.',
		'Chat-Agents sind auf diesem Bulle-Server nicht aktiviert.'
	],
	'matrix.notCreated': ['Not created yet', 'Pas encore créé', 'Aún no creado', 'Noch nicht erstellt'],
	'matrix.noChannels': [
		'No channels yet. Create a space in Prisme and the agent joins it within ten minutes.',
		'Aucun canal pour l’instant. Créez un espace dans Prisme : l’agent le rejoint en moins de dix minutes.',
		'Aún no hay canales. Crea un espacio en Prisme y el agente se une en menos de diez minutos.',
		'Noch keine Kanäle. Erstellen Sie einen Bereich in Prisme, und der Agent tritt innerhalb von zehn Minuten bei.'
	],

	'models.title': ['Default model', 'Modèle par défaut', 'Modelo predeterminado', 'Standardmodell'],
	'models.description': [
		"The model a new session starts on, here and in the CLI. The list is what the current provider offers; the CLI's /model browses the same catalogue.",
		'Le modèle sur lequel démarre une nouvelle session, ici et dans la CLI. La liste reprend l’offre du fournisseur actuel ; /model dans la CLI parcourt le même catalogue.',
		'El modelo con el que empieza una sesión nueva, aquí y en la CLI. La lista es la oferta del proveedor actual; /model en la CLI recorre el mismo catálogo.',
		'Das Modell, mit dem eine neue Sitzung startet, hier und in der CLI. Die Liste zeigt das Angebot des aktuellen Anbieters; /model in der CLI durchsucht denselben Katalog.'
	],
	'models.loadFailed': [
		'Could not load the model settings',
		'Impossible de charger les réglages du modèle',
		'No se pudieron cargar los ajustes del modelo',
		'Modelleinstellungen konnten nicht geladen werden'
	],
	'models.saved': [
		'Default model saved. New sessions start on it.',
		'Modèle par défaut enregistré. Les nouvelles sessions démarrent avec lui.',
		'Modelo predeterminado guardado. Las sesiones nuevas empiezan con él.',
		'Standardmodell gespeichert. Neue Sitzungen starten damit.'
	],
	'models.saveFailed': [
		'Could not save the default model',
		'Impossible d’enregistrer le modèle par défaut',
		'No se pudo guardar el modelo predeterminado',
		'Standardmodell konnte nicht gespeichert werden'
	],
	'models.current': ['Current:', 'Actuel :', 'Actual:', 'Aktuell:'],
	'models.serverDefault': [
		'server default ({model})',
		'défaut du serveur ({model})',
		'predeterminado del servidor ({model})',
		'Serverstandard ({model})'
	],
	'models.searchLabel': ['Search models', 'Rechercher des modèles', 'Buscar modelos', 'Modelle suchen'],
	'models.count': [
		'{count} models from this provider',
		'{count} modèles chez ce fournisseur',
		'{count} modelos de este proveedor',
		'{count} Modelle dieses Anbieters'
	],
	'models.searchPlaceholder': [
		'Search by name or id, e.g. mimo, claude, gpt',
		'Rechercher par nom ou id, ex. mimo, claude, gpt',
		'Buscar por nombre o id, p. ej. mimo, claude, gpt',
		'Nach Name oder ID suchen, z. B. mimo, claude, gpt'
	],
	'models.noMatch': [
		'No model matches “{query}”.',
		'Aucun modèle ne correspond à « {query} ».',
		'Ningún modelo coincide con «{query}».',
		'Kein Modell passt zu „{query}“.'
	],
	'models.showing': [
		'Showing {shown} of {total}. Narrow the search to see the rest.',
		'{shown} affichés sur {total}. Affinez la recherche pour voir le reste.',
		'Mostrando {shown} de {total}. Afina la búsqueda para ver el resto.',
		'{shown} von {total} angezeigt. Verfeinern Sie die Suche, um den Rest zu sehen.'
	],
	'models.useDefault': [
		'Use the server default',
		'Utiliser le défaut du serveur',
		'Usar el predeterminado del servidor',
		'Serverstandard verwenden'
	],
	'models.save': ['Save model', 'Enregistrer le modèle', 'Guardar modelo', 'Modell speichern'],

	'providers.title': ['AI providers', 'Fournisseurs d’IA', 'Proveedores de IA', 'KI-Anbieter'],
	'providers.description': [
		'Keys are stored encrypted on the server. You can replace or remove a key, never read it back.',
		'Les clés sont stockées chiffrées sur le serveur. Vous pouvez remplacer ou supprimer une clé, jamais la relire.',
		'Las claves se guardan cifradas en el servidor. Se pueden reemplazar o eliminar, pero nunca volver a leer.',
		'Schlüssel werden verschlüsselt auf dem Server gespeichert. Sie können einen Schlüssel ersetzen oder entfernen, aber nie wieder auslesen.'
	],
	'providers.loadFailed': [
		'Could not load provider keys',
		'Impossible de charger les clés des fournisseurs',
		'No se pudieron cargar las claves de los proveedores',
		'Anbieterschlüssel konnten nicht geladen werden'
	],
	'providers.requestFailed': ['Request failed', 'Échec de la requête', 'Error en la solicitud', 'Anfrage fehlgeschlagen'],
	'providers.source.user': ['Your key', 'Votre clé', 'Clave propia', 'Ihr Schlüssel'],
	'providers.source.instance': ['Instance key', 'Clé de l’instance', 'Clave de la instancia', 'Instanzschlüssel'],
	'providers.source.none': ['No key', 'Aucune clé', 'Sin clave', 'Kein Schlüssel'],
	'providers.saved': ['{provider} key saved', 'Clé {provider} enregistrée', 'Clave de {provider} guardada', '{provider}-Schlüssel gespeichert'],
	'providers.removed': ['{provider} key removed', 'Clé {provider} supprimée', 'Clave de {provider} eliminada', '{provider}-Schlüssel entfernt'],
	'providers.works': ['{provider} key works', 'La clé {provider} fonctionne', 'La clave de {provider} funciona', '{provider}-Schlüssel funktioniert'],
	'providers.testFailed': [
		'{provider} key failed the test',
		'La clé {provider} a échoué au test',
		'La clave de {provider} no superó la prueba',
		'{provider}-Schlüssel hat den Test nicht bestanden'
	],
	'providers.keyAria': ['{provider} API key', 'Clé d’API {provider}', 'Clave de API de {provider}', '{provider}-API-Schlüssel'],
	'providers.pasteReplace': [
		'Paste a new key to replace it',
		'Collez une nouvelle clé pour la remplacer',
		'Pega una clave nueva para reemplazarla',
		'Neuen Schlüssel zum Ersetzen einfügen'
	],
	'providers.paste': ['Paste an API key', 'Collez une clé d’API', 'Pega una clave de API', 'API-Schlüssel einfügen'],
	'providers.test': ['Test', 'Tester', 'Probar', 'Testen'],
	'providers.remove': ['Remove', 'Supprimer', 'Eliminar', 'Entfernen'],

	'notifications.title': ['Notifications', 'Notifications', 'Notificaciones', 'Benachrichtigungen'],
	'notifications.description': [
		'Get a notification when a job finishes or your assistant needs an approval.',
		'Recevez une notification quand une tâche se termine ou que votre assistant attend une validation.',
		'Recibe una notificación cuando termine una tarea o el asistente necesite una aprobación.',
		'Erhalten Sie eine Benachrichtigung, wenn eine Aufgabe fertig ist oder Ihr Assistent eine Freigabe braucht.'
	],
	'notifications.on': ['Notifications are on', 'Notifications activées', 'Notificaciones activadas', 'Benachrichtigungen sind aktiviert'],
	'notifications.turnOnFailed': [
		'Could not turn on notifications',
		'Impossible d’activer les notifications',
		'No se pudieron activar las notificaciones',
		'Benachrichtigungen konnten nicht aktiviert werden'
	],
	'notifications.testSent': [
		'Test notification sent',
		'Notification de test envoyée',
		'Notificación de prueba enviada',
		'Testbenachrichtigung gesendet'
	],
	'notifications.testFailed': [
		'Could not send the test',
		'Impossible d’envoyer le test',
		'No se pudo enviar la prueba',
		'Test konnte nicht gesendet werden'
	],
	'notifications.linkCopied': ['Link copied', 'Lien copié', 'Enlace copiado', 'Link kopiert'],
	'notifications.copyFailed': [
		'Could not copy the link',
		'Impossible de copier le lien',
		'No se pudo copiar el enlace',
		'Link konnte nicht kopiert werden'
	],
	'notifications.notSetUpNative': [
		'Native notifications are not set up on this server yet.',
		'Les notifications natives ne sont pas encore configurées sur ce serveur.',
		'Las notificaciones nativas aún no están configuradas en este servidor.',
		'Native Benachrichtigungen sind auf diesem Server noch nicht eingerichtet.'
	],
	'notifications.notSetUpBrowser': [
		'Browser notifications are not set up on this server yet.',
		'Les notifications du navigateur ne sont pas encore configurées sur ce serveur.',
		'Las notificaciones del navegador aún no están configuradas en este servidor.',
		'Browser-Benachrichtigungen sind auf diesem Server noch nicht eingerichtet.'
	],
	'notifications.unsupported': [
		'This browser cannot show notifications.',
		'Ce navigateur ne peut pas afficher de notifications.',
		'Este navegador no puede mostrar notificaciones.',
		'Dieser Browser kann keine Benachrichtigungen anzeigen.'
	],
	'notifications.blockedDesktop': [
		'Notifications are blocked. Allow Bulle in your system notification settings.',
		'Les notifications sont bloquées. Autorisez Bulle dans les réglages de notification du système.',
		'Las notificaciones están bloqueadas. Permite Bulle en los ajustes de notificaciones del sistema.',
		'Benachrichtigungen sind blockiert. Erlauben Sie Bulle in den Benachrichtigungseinstellungen des Systems.'
	],
	'notifications.blockedNative': [
		'Notifications are blocked. Turn them on in Settings > Apps > Bulle > Notifications.',
		'Les notifications sont bloquées. Activez-les dans Réglages > Applications > Bulle > Notifications.',
		'Las notificaciones están bloqueadas. Actívalas en Ajustes > Aplicaciones > Bulle > Notificaciones.',
		'Benachrichtigungen sind blockiert. Aktivieren Sie sie unter Einstellungen > Apps > Bulle > Benachrichtigungen.'
	],
	'notifications.blockedBrowser': [
		'Notifications are blocked for this site. Allow them in your browser settings.',
		'Les notifications sont bloquées pour ce site. Autorisez-les dans les réglages du navigateur.',
		'Las notificaciones están bloqueadas para este sitio. Permítelas en los ajustes del navegador.',
		'Benachrichtigungen sind für diese Website blockiert. Erlauben Sie sie in den Browsereinstellungen.'
	],
	'notifications.turnOn': [
		'Turn on notifications',
		'Activer les notifications',
		'Activar notificaciones',
		'Benachrichtigungen aktivieren'
	],
	'notifications.onForDevice': [
		'Notifications are on for this device.',
		'Les notifications sont activées sur cet appareil.',
		'Las notificaciones están activadas en este dispositivo.',
		'Benachrichtigungen sind für dieses Gerät aktiviert.'
	],
	'notifications.sendTest': ['Send test', 'Envoyer un test', 'Enviar prueba', 'Test senden'],
	'notifications.ntfyHint': [
		'You can also follow your private ntfy topic from the ntfy app.',
		'Vous pouvez aussi suivre votre sujet ntfy privé depuis l’application ntfy.',
		'También puedes seguir tu tema privado de ntfy desde la app de ntfy.',
		'Sie können Ihr privates ntfy-Thema auch in der ntfy-App verfolgen.'
	],
	'notifications.server': ['Server', 'Serveur', 'Servidor', 'Server'],
	'notifications.topic': ['Topic', 'Sujet', 'Tema', 'Thema'],
	'notifications.openNtfy': ['Open in ntfy', 'Ouvrir dans ntfy', 'Abrir en ntfy', 'In ntfy öffnen'],
	'notifications.copyLink': ['Copy subscribe link', 'Copier le lien d’abonnement', 'Copiar enlace de suscripción', 'Abo-Link kopieren'],
	'notifications.tabOnly': [
		'This browser has no push service, so Bulle notifies you only while a tab is open. Install the desktop app for notifications when it is closed.',
		'Ce navigateur n’a pas de service push : Bulle ne vous notifie que tant qu’un onglet est ouvert. Installez l’application de bureau pour être notifié même fermée.',
		'Este navegador no tiene servicio push, así que Bulle solo notifica mientras haya una pestaña abierta. Instala la aplicación de escritorio para recibir notificaciones con ella cerrada.',
		'Dieser Browser hat keinen Push-Dienst, daher benachrichtigt Bulle nur bei geöffnetem Tab. Installieren Sie die Desktop-App, um auch bei geschlossener App benachrichtigt zu werden.'
	],
	'notifications.pushServiceOff': [
		'This browser has no push service turned on. In Brave, enable "Use Google services for push messaging" in brave://settings/privacy and restart it. Helium and other Chromium builds without Google services cannot receive web push: use Chrome, Firefox or the Bulle app instead.',
		'Ce navigateur n’a aucun service push activé. Dans Brave, activez « Utiliser les services Google pour la messagerie push » dans brave://settings/privacy puis redémarrez-le. Helium et les autres versions de Chromium sans services Google ne reçoivent pas le push web : utilisez Chrome, Firefox ou l’application Bulle.',
		'Este navegador no tiene ningún servicio push activado. En Brave, activa «Usar servicios de Google para mensajería push» en brave://settings/privacy y reinícialo. Helium y otras versiones de Chromium sin servicios de Google no pueden recibir push web: usa Chrome, Firefox o la aplicación de Bulle.',
		'In diesem Browser ist kein Push-Dienst aktiviert. Aktivieren Sie in Brave unter brave://settings/privacy „Google-Dienste für Push-Nachrichten verwenden“ und starten Sie ihn neu. Helium und andere Chromium-Builds ohne Google-Dienste können kein Web-Push empfangen: Nutzen Sie Chrome, Firefox oder die Bulle-App.'
	],
	'notifications.channelDescription': [
		'Approvals and finished jobs',
		'Validations et tâches terminées',
		'Aprobaciones y tareas terminadas',
		'Freigaben und abgeschlossene Aufgaben'
	],
	'notifications.prompt.text': [
		'Get a notification when a job is ready.',
		'Recevez une notification quand une tâche est prête.',
		'Recibe una notificación cuando una tarea esté lista.',
		'Erhalten Sie eine Benachrichtigung, wenn eine Aufgabe fertig ist.'
	],
	'notifications.prompt.turnOn': ['Turn on', 'Activer', 'Activar', 'Aktivieren'],

	'spaces.title': ['Spaces', 'Espaces', 'Espacios', 'Bereiche'],
	'spaces.description': [
		'Shared spaces follow your Facile spaces automatically, and the assistant uses the matching Opus workspace and Mycelium space in each. Manage members in Registre.',
		'Les espaces partagés suivent automatiquement vos espaces Facile, et l’assistant utilise l’espace de travail Opus et l’espace Mycelium correspondants dans chacun. Gérez les membres dans Registre.',
		'Los espacios compartidos siguen automáticamente tus espacios de Facile, y el asistente usa en cada uno el espacio de trabajo de Opus y el espacio de Mycelium correspondientes. Gestiona los miembros en Registre.',
		'Gemeinsame Bereiche folgen automatisch Ihren Facile-Bereichen, und der Assistent nutzt in jedem den passenden Opus-Arbeitsbereich und Mycelium-Bereich. Mitglieder verwalten Sie in Registre.'
	],
	'spaces.new': ['New space', 'Nouvel espace', 'Nuevo espacio', 'Neuer Bereich'],
	'spaces.nameAria': ['Space name', 'Nom de l’espace', 'Nombre del espacio', 'Name des Bereichs'],
	'spaces.save': ['Save', 'Enregistrer', 'Guardar', 'Speichern'],
	'spaces.current': ['Current', 'Actuel', 'Actual', 'Aktuell'],
	'spaces.shared': ['Shared', 'Partagé', 'Compartido', 'Geteilt'],
	'spaces.rename': ['Rename', 'Renommer', 'Renombrar', 'Umbenennen'],
	'spaces.delete': ['Delete', 'Supprimer', 'Eliminar', 'Löschen'],
	'spaces.renamed': ['Space renamed', 'Espace renommé', 'Espacio renombrado', 'Bereich umbenannt'],
	'spaces.renameFailed': [
		'Could not rename the space',
		'Impossible de renommer l’espace',
		'No se pudo renombrar el espacio',
		'Bereich konnte nicht umbenannt werden'
	],
	'spaces.deleted': ['{name} deleted', '{name} supprimé', '{name} eliminado', '{name} gelöscht'],
	'spaces.deleteFailed': [
		'Could not delete the space',
		'Impossible de supprimer l’espace',
		'No se pudo eliminar el espacio',
		'Bereich konnte nicht gelöscht werden'
	],
	'spaces.deleteTitle': ['Delete {name}?', 'Supprimer {name} ?', '¿Eliminar {name}?', '{name} löschen?'],
	'spaces.deleteTitleFallback': [
		'Delete this space?',
		'Supprimer cet espace ?',
		'¿Eliminar este espacio?',
		'Diesen Bereich löschen?'
	],
	'spaces.deleteDescription': [
		'Every chat and job in this space is deleted for good.',
		'Toutes les discussions et tâches de cet espace sont supprimées définitivement.',
		'Todos los chats y tareas de este espacio se eliminan de forma definitiva.',
		'Alle Chats und Aufgaben in diesem Bereich werden endgültig gelöscht.'
	],
	'spaces.deleteConfirm': ['Delete space', 'Supprimer l’espace', 'Eliminar espacio', 'Bereich löschen'],
	'spaces.create.title': ['New space', 'Nouvel espace', 'Nuevo espacio', 'Neuer Bereich'],
	'spaces.create.intro': [
		'A space keeps its own chats, jobs and default accounts apart from your other spaces.',
		'Un espace garde ses discussions, ses tâches et ses comptes par défaut séparés de vos autres espaces.',
		'Un espacio mantiene sus chats, tareas y cuentas predeterminadas separados de tus otros espacios.',
		'Ein Bereich hält seine Chats, Aufgaben und Standardkonten getrennt von Ihren anderen Bereichen.'
	],
	'spaces.create.name': ['Name', 'Nom', 'Nombre', 'Name'],
	'spaces.create.placeholder': ['e.g. Work', 'ex. Travail', 'p. ej. Trabajo', 'z. B. Arbeit'],
	'spaces.create.failed': [
		'Could not create the space',
		'Impossible de créer l’espace',
		'No se pudo crear el espacio',
		'Bereich konnte nicht erstellt werden'
	],
	'spaces.create.creating': ['Creating...', 'Création…', 'Creando…', 'Wird erstellt…'],
	'spaces.create.submit': ['Create space', 'Créer l’espace', 'Crear espacio', 'Bereich erstellen'],

	'updater.title': ['Updates', 'Mises à jour', 'Actualizaciones', 'Updates'],
	'updater.description': [
		'Check for new Bulle desktop releases and install them.',
		'Recherchez les nouvelles versions de Bulle pour ordinateur et installez-les.',
		'Busca nuevas versiones de Bulle para escritorio e instálalas.',
		'Suchen Sie nach neuen Bulle-Desktop-Versionen und installieren Sie sie.'
	],
	'updater.descriptionAndroid': [
		'Check for a new Bulle for Android and download the APK.',
		'Recherchez une nouvelle version de Bulle pour Android et téléchargez l’APK.',
		'Busca una nueva versión de Bulle para Android y descarga el APK.',
		'Suchen Sie nach einer neuen Bulle-Version für Android und laden Sie die APK herunter.'
	],
	'updater.downloadApk': ['Download the APK', 'Télécharger l’APK', 'Descargar el APK', 'APK herunterladen'],
	'updater.checking': [
		'Checking for updates…',
		'Recherche de mises à jour…',
		'Buscando actualizaciones…',
		'Suche nach Updates…'
	],
	'updater.available': [
		'Update v{version} available',
		'Mise à jour v{version} disponible',
		'Actualización v{version} disponible',
		'Update v{version} verfügbar'
	],
	'updater.downloadingPercent': [
		'Downloading {percent}%',
		'Téléchargement {percent} %',
		'Descargando {percent} %',
		'Download {percent} %'
	],
	'updater.ready': [
		'Update ready to install',
		'Mise à jour prête à installer',
		'Actualización lista para instalar',
		'Update bereit zur Installation'
	],
	'updater.failed': ['Update failed', 'Échec de la mise à jour', 'Error en la actualización', 'Update fehlgeschlagen'],
	'updater.upToDate': ['Bulle is up to date', 'Bulle est à jour', 'Bulle está actualizado', 'Bulle ist aktuell'],
	'updater.idle': ['Ready to check', 'Prêt à vérifier', 'Listo para buscar', 'Bereit zur Prüfung'],
	'updater.currentBuild': ['Current release build', 'Version publiée actuelle', 'Versión publicada actual', 'Aktueller Release-Build'],
	'updater.install': ['Download & Install', 'Télécharger et installer', 'Descargar e instalar', 'Herunterladen und installieren'],
	'updater.restart': ['Restart & Install', 'Redémarrer et installer', 'Reiniciar e instalar', 'Neu starten und installieren'],
	'updater.checkingButton': ['Checking…', 'Recherche…', 'Buscando…', 'Prüfung…'],
	'updater.check': ['Check for updates', 'Rechercher des mises à jour', 'Buscar actualizaciones', 'Nach Updates suchen'],
	'updater.downloading': [
		'Downloading update…',
		'Téléchargement de la mise à jour…',
		'Descargando la actualización…',
		'Update wird heruntergeladen…'
	],
	'updater.changelog': [
		'View full changelog and release history',
		'Voir le journal des modifications et l’historique des versions',
		'Ver el registro de cambios y el historial de versiones',
		'Vollständiges Änderungsprotokoll und Versionsverlauf ansehen'
	],
	'updater.releaseHistory': ['Release history', 'Historique des versions', 'Historial de versiones', 'Versionsverlauf'],
	'updater.noUpdate': [
		'No update available to install',
		'Aucune mise à jour à installer',
		'No hay ninguna actualización para instalar',
		'Kein Update zur Installation verfügbar'
	],

	'usage.title': ['Usage', 'Utilisation', 'Uso', 'Nutzung'],
	'usage.description': [
		'Tokens and cost of your conversations and background jobs.',
		'Jetons et coût de vos conversations et de vos tâches de fond.',
		'Tokens y coste de tus conversaciones y tareas en segundo plano.',
		'Tokens und Kosten Ihrer Unterhaltungen und Hintergrundaufgaben.'
	],
	'usage.today': ['Today', 'Aujourd’hui', 'Hoy', 'Heute'],
	'usage.week': ['Last 7 days', '7 derniers jours', 'Últimos 7 días', 'Letzte 7 Tage'],
	'usage.month': ['Last 30 days', '30 derniers jours', 'Últimos 30 días', 'Letzte 30 Tage'],
	'usage.cost': ['Cost', 'Coût', 'Coste', 'Kosten'],
	'usage.runs': [
		{ one: '{count} run', other: '{count} runs' },
		{ one: '{count} exécution', other: '{count} exécutions' },
		{ one: '{count} ejecución', other: '{count} ejecuciones' },
		{ one: '{count} Ausführung', other: '{count} Ausführungen' }
	],
	'usage.tokens': ['{count} tokens', '{count} jetons', '{count} tokens', '{count} Tokens'],
	'usage.delta': ['{tokens} · {runs}', '{tokens} · {runs}', '{tokens} · {runs}', '{tokens} · {runs}'],
	'usage.loadFailed': [
		'Could not load usage',
		'Impossible de charger l’utilisation',
		'No se pudo cargar el uso',
		'Nutzung konnte nicht geladen werden'
	],
	'usage.nothing': [
		'Nothing in the last 30 days.',
		'Rien sur les 30 derniers jours.',
		'Nada en los últimos 30 días.',
		'Nichts in den letzten 30 Tagen.'
	],
	'usage.daily': ['Daily cost', 'Coût quotidien', 'Coste diario', 'Tägliche Kosten'],
	'usage.noRuns': ['No runs yet', 'Aucune exécution pour l’instant', 'Aún no hay ejecuciones', 'Noch keine Ausführungen'],
	'usage.input': ['Input', 'Entrée', 'Entrada', 'Eingabe'],
	'usage.output': ['Output', 'Sortie', 'Salida', 'Ausgabe'],
	'usage.cacheRead': ['Cache read', 'Lecture du cache', 'Lectura de caché', 'Cache-Lesezugriffe'],
	'usage.cacheWrite': ['Cache write', 'Écriture du cache', 'Escritura de caché', 'Cache-Schreibzugriffe'],
	'usage.byModel': ['By model', 'Par modèle', 'Por modelo', 'Nach Modell'],
	'usage.byChannel': ['Chat and jobs', 'Discussions et tâches', 'Chat y tareas', 'Chat und Aufgaben'],
	'usage.footer': [
		'Last 30 days, in {timezone} time.',
		'30 derniers jours, en heure de {timezone}.',
		'Últimos 30 días, en hora de {timezone}.',
		'Letzte 30 Tage, in {timezone}-Zeit.'
	],

	'suite.title': ['Connected apps', 'Applications connectées', 'Aplicaciones conectadas', 'Verbundene Apps'],
	'suite.description': [
		"Your assistant reads your own data in these Facile apps. It never sees anyone else's.",
		'Votre assistant lit vos propres données dans ces applications Facile. Il ne voit jamais celles des autres.',
		'El asistente lee tus propios datos en estas aplicaciones de Facile. Nunca ve los de nadie más.',
		'Ihr Assistent liest Ihre eigenen Daten in diesen Facile-Apps. Die Daten anderer sieht er nie.'
	],
	'suite.loadFailed': [
		'Could not load connected apps',
		'Impossible de charger les applications connectées',
		'No se pudieron cargar las aplicaciones conectadas',
		'Verbundene Apps konnten nicht geladen werden'
	],
	'suite.openOnce': [
		'Open {name} once to create your account',
		'Ouvrez {name} une fois pour créer votre compte',
		'Abre {name} una vez para crear tu cuenta',
		'Öffnen Sie {name} einmal, um Ihr Konto anzulegen'
	],
	'suite.signInAgain': ['Sign in again', 'Se reconnecter', 'Iniciar sesión de nuevo', 'Erneut anmelden'],
	'suite.state.connected': ['Connected', 'Connectée', 'Conectada', 'Verbunden'],
	'suite.state.no_account': ['No account yet', 'Pas encore de compte', 'Aún sin cuenta', 'Noch kein Konto'],
	'suite.state.reconnect': ['Reconnect', 'À reconnecter', 'Reconectar', 'Neu verbinden'],
	'suite.state.unavailable': ['Unavailable', 'Indisponible', 'No disponible', 'Nicht verfügbar'],

	'connect.title': [
		'Connect Mycelium to keep artifacts',
		'Connectez Mycelium pour conserver les artefacts',
		'Conecta Mycelium para conservar los artefactos',
		'Mycelium verbinden, um Artefakte zu behalten'
	],
	'connect.description': [
		'Artifacts and memory live in Mycelium. Connect it in Settings, Apps, and the assistant can save documents you can come back to.',
		'Les artefacts et la mémoire vivent dans Mycelium. Connectez-le dans Réglages, Applications : l’assistant pourra enregistrer des documents que vous retrouverez.',
		'Los artefactos y la memoria viven en Mycelium. Conéctalo en Ajustes, Aplicaciones, y el asistente podrá guardar documentos a los que volver.',
		'Artefakte und Gedächtnis liegen in Mycelium. Verbinden Sie es unter Einstellungen, Apps, und der Assistent kann Dokumente speichern, zu denen Sie zurückkehren können.'
	],
	'connect.open': [
		'Open Settings, Apps',
		'Ouvrir Réglages, Applications',
		'Abrir Ajustes, Aplicaciones',
		'Einstellungen, Apps öffnen'
	]
} as const satisfies Record<string, Row>;
