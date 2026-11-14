import type { Row } from '../row';

export const shell = {
	'common.cancel': ['Cancel', 'Annuler', 'Cancelar', 'Abbrechen'],
	'common.save': ['Save', 'Enregistrer', 'Guardar', 'Speichern'],
	'common.delete': ['Delete', 'Supprimer', 'Eliminar', 'Löschen'],
	'common.close': ['Close', 'Fermer', 'Cerrar', 'Schließen'],
	'common.retry': ['Retry', 'Réessayer', 'Reintentar', 'Erneut versuchen'],
	'common.tryAgain': ['Try again', 'Réessayer', 'Intentar de nuevo', 'Erneut versuchen'],
	'common.dismiss': ['Dismiss', 'Ignorer', 'Descartar', 'Verwerfen'],
	'common.loading': ['Loading', 'Chargement', 'Cargando', 'Wird geladen'],
	'common.open': ['Open', 'Ouvrir', 'Abrir', 'Öffnen'],
	'common.copy': ['Copy', 'Copier', 'Copiar', 'Kopieren'],
	'common.copied': ['Copied', 'Copié', 'Copiado', 'Kopiert'],
	'error.requestFailed': [
		'Request failed with status {status}',
		'La requête a échoué avec le statut {status}',
		'La solicitud falló con el estado {status}',
		'Anfrage mit Status {status} fehlgeschlagen'
	],
	'error.requestFailedShort': ['The request failed', 'La requête a échoué', 'La solicitud falló', 'Die Anfrage ist fehlgeschlagen'],
	'error.unreadable': [
		'The server sent an unreadable answer',
		'Le serveur a envoyé une réponse illisible',
		'El servidor envió una respuesta ilegible',
		'Der Server hat eine unlesbare Antwort gesendet'
	],
	'error.uploadCancelled': ['The upload was cancelled', 'Le téléversement a été annulé', 'La subida se canceló', 'Der Upload wurde abgebrochen'],
	'error.uploadUnreachable': [
		'The upload could not reach the server',
		'Le téléversement n’a pas atteint le serveur',
		'La subida no pudo llegar al servidor',
		'Der Upload konnte den Server nicht erreichen'
	],
	'time.now': ['now', 'maintenant', 'ahora', 'jetzt'],
	'nav.ideas': ['Ideas', 'Idées', 'Ideas', 'Ideen'],
	'nav.artifacts': ['Artifacts', 'Artefacts', 'Artefactos', 'Artefakte'],
	'sidebar.collapse': ['Collapse the sidebar', 'Réduire la barre latérale', 'Contraer la barra lateral', 'Seitenleiste einklappen'],
	'sidebar.collapseHint': [
		'Collapse the sidebar (Ctrl+B)',
		'Réduire la barre latérale (Ctrl+B)',
		'Contraer la barra lateral (Ctrl+B)',
		'Seitenleiste einklappen (Strg+B)'
	],
	'sidebar.open': ['Open the sidebar', 'Ouvrir la barre latérale', 'Abrir la barra lateral', 'Seitenleiste öffnen'],
	'sidebar.close': ['Close the sidebar', 'Fermer la barre latérale', 'Cerrar la barra lateral', 'Seitenleiste schließen'],
	'sidebar.navigation': ['Navigation', 'Navigation', 'Navegación', 'Navigation'],
	'sidebar.pages': ['Pages', 'Pages', 'Páginas', 'Seiten'],
	'sidebar.running': ['Running', 'En cours', 'En curso', 'Läuft'],
	'sidebar.newCount': ['{count} new', '{count} nouvelles', '{count} nuevas', '{count} neu'],
	'profile.jobs': ['Jobs', 'Tâches', 'Tareas', 'Aufgaben'],
	'profile.getApps': ['Get the apps', 'Obtenir les applications', 'Obtener las aplicaciones', 'Apps herunterladen'],
	'profile.getDesktop': ['Get the desktop app', 'Obtenir l’application de bureau', 'Obtener la aplicación de escritorio', 'Desktop-App herunterladen'],
	'profile.jobRunning': ['A job is running', 'Une tâche est en cours', 'Hay una tarea en curso', 'Eine Aufgabe läuft'],
	'app.unreachableTitle': ['Bulle is not answering', 'Bulle ne répond pas', 'Bulle no responde', 'Bulle antwortet nicht'],
	'app.unreachableBody': [
		'The server did not return your account. You are still signed in.',
		'Le serveur n’a pas renvoyé votre compte. Vous êtes toujours connecté.',
		'El servidor no devolvió tu cuenta. Sigues con la sesión iniciada.',
		'Der Server hat Ihr Konto nicht zurückgegeben. Sie sind weiterhin angemeldet.'
	],
	'session.working': ['Working', 'En cours', 'Trabajando', 'Arbeitet'],
	'session.needsApproval': ['Needs approval', 'Attend une validation', 'Requiere aprobación', 'Wartet auf Freigabe'],
	'session.list': ['Sessions', 'Sessions', 'Sesiones', 'Sitzungen'],
	'session.unread': ['Unread', 'Non lu', 'No leído', 'Ungelesen'],
	'session.one': ['session', 'session', 'sesión', 'Sitzung'],
	'session.loadFailed': [
		'Could not load sessions',
		'Impossible de charger les sessions',
		'No se pudieron cargar las sesiones',
		'Sitzungen konnten nicht geladen werden'
	],
	'session.deleteFailed': [
		'Could not delete the session',
		'Impossible de supprimer la session',
		'No se pudo eliminar la sesión',
		'Die Sitzung konnte nicht gelöscht werden'
	],
	'session.emptyTitle': ['No sessions yet', 'Aucune session pour l’instant', 'Aún no hay sesiones', 'Noch keine Sitzungen'],
	'session.emptyBody': [
		'Start a chat and it will show up here.',
		'Lancez une discussion, elle apparaîtra ici.',
		'Inicia un chat y aparecerá aquí.',
		'Starten Sie einen Chat, er erscheint hier.'
	],
	'session.delete': ['Delete {title}', 'Supprimer {title}', 'Eliminar {title}', '{title} löschen'],
	'session.deleteTitle': ['Delete this session?', 'Supprimer cette session ?', '¿Eliminar esta sesión?', 'Diese Sitzung löschen?'],
	'session.deleteBody': [
		'The conversation and its history are removed for good.',
		'La conversation et son historique sont supprimés définitivement.',
		'La conversación y su historial se eliminan para siempre.',
		'Die Unterhaltung und ihr Verlauf werden endgültig gelöscht.'
	],
	'search.placeholder': ['Search chats', 'Rechercher des discussions', 'Buscar chats', 'Chats durchsuchen'],
	'search.label': ['Search past chats', 'Rechercher dans les discussions passées', 'Buscar en chats anteriores', 'Frühere Chats durchsuchen'],
	'search.searching': ['Searching', 'Recherche en cours', 'Buscando', 'Suche läuft'],
	'search.results': ['Search results', 'Résultats de la recherche', 'Resultados de la búsqueda', 'Suchergebnisse'],
	'search.none': [
		'No chat matches "{query}".',
		'Aucune discussion ne correspond à « {query} ».',
		'Ningún chat coincide con «{query}».',
		'Kein Chat passt zu „{query}“.'
	],
	'search.unavailable': [
		'Search is not available right now.',
		'La recherche n’est pas disponible pour le moment.',
		'La búsqueda no está disponible ahora.',
		'Die Suche ist derzeit nicht verfügbar.'
	],
	'assistant.archetype.pragmatic_senior.label': ['Pragmatic Senior', 'Senior pragmatique', 'Senior pragmático', 'Pragmatischer Senior'],
	'assistant.archetype.pragmatic_senior.description': [
		'Direct, battle-tested, clean code first.',
		'Direct, éprouvé, le code propre d’abord.',
		'Directo, probado en batalla, código limpio primero.',
		'Direkt, praxiserprobt, sauberer Code zuerst.'
	],
	'assistant.archetype.terse_minimalist.label': ['Terse Minimalist', 'Minimaliste concis', 'Minimalista conciso', 'Knapper Minimalist'],
	'assistant.archetype.terse_minimalist.description': [
		'Ultra concise, code-heavy, minimum words.',
		'Ultra concis, beaucoup de code, le moins de mots possible.',
		'Ultraconciso, mucho código, las mínimas palabras.',
		'Extrem knapp, viel Code, so wenige Worte wie möglich.'
	],
	'assistant.archetype.supportive_mentor.label': ['Supportive Mentor', 'Mentor bienveillant', 'Mentor atento', 'Unterstützender Mentor'],
	'assistant.archetype.supportive_mentor.description': [
		'Thorough, clear explanations, patient and insightful.',
		'Explications claires et complètes, patient et perspicace.',
		'Explicaciones claras y completas, paciente y perspicaz.',
		'Gründliche, klare Erklärungen, geduldig und einfühlsam.'
	],
	'assistant.archetype.witty_hacker.label': ['Witty Hacker', 'Hacker malin', 'Hacker ingenioso', 'Witziger Hacker'],
	'assistant.archetype.witty_hacker.description': [
		'Clever, lightly sarcastic, highly technical, loves elegant hacks.',
		'Astucieux, légèrement sarcastique, très technique, aime les astuces élégantes.',
		'Astuto, ligeramente sarcástico, muy técnico, ama los trucos elegantes.',
		'Clever, leicht sarkastisch, sehr technisch, liebt elegante Kniffe.'
	],
	'assistant.trait.concise': ['Concise', 'Concis', 'Conciso', 'Prägnant'],
	'assistant.trait.explain_reasoning': ['Explain Reasoning', 'Expliquer le raisonnement', 'Explicar el razonamiento', 'Begründungen erklären'],
	'assistant.trait.witty': ['Light Sarcasm', 'Sarcasme léger', 'Sarcasmo ligero', 'Leichter Sarkasmus'],
	'assistant.trait.strict': ['Strict Type Safety', 'Typage strict', 'Tipado estricto', 'Strikte Typsicherheit'],
	'assistant.trait.creative': ['Creative Solutions', 'Solutions créatives', 'Soluciones creativas', 'Kreative Lösungen'],
	'ble.connecting': ['Connecting to your Pastille', 'Connexion à votre Pastille', 'Conectando con tu Pastille', 'Verbindung zu Ihrer Pastille'],
	'ble.connectingTo': ['Connecting to {name}', 'Connexion à {name}', 'Conectando con {name}', 'Verbindung zu {name}'],
	'ble.tap': [
		'Tap your Pastille to confirm pairing',
		'Touchez votre Pastille pour confirmer l’appairage',
		'Toca tu Pastille para confirmar el emparejamiento',
		'Tippen Sie auf Ihre Pastille, um das Koppeln zu bestätigen'
	],
	'ble.adding': ['Adding it to your account', 'Ajout à votre compte', 'Añadiéndola a tu cuenta', 'Wird Ihrem Konto hinzugefügt'],
	'ble.paired': ['Paired', 'Appairée', 'Emparejada', 'Gekoppelt'],
	'ble.wrongDevice': [
		'The Pastille you picked is not {name}.',
		'La Pastille choisie n’est pas {name}.',
		'La Pastille que elegiste no es {name}.',
		'Die gewählte Pastille ist nicht {name}.'
	],
	'ble.noBluetooth': [
		'This browser has no Web Bluetooth. Use Chrome or Edge.',
		'Ce navigateur ne prend pas en charge Web Bluetooth. Utilisez Chrome ou Edge.',
		'Este navegador no admite Web Bluetooth. Usa Chrome o Edge.',
		'Dieser Browser unterstützt kein Web Bluetooth. Verwenden Sie Chrome oder Edge.'
	],
	'ble.noneSelected': ['No Pastille was selected.', 'Aucune Pastille n’a été choisie.', 'No se eligió ninguna Pastille.', 'Es wurde keine Pastille ausgewählt.'],
	'ble.noGatt': [
		'This Pastille has no GATT server.',
		'Cette Pastille n’a pas de serveur GATT.',
		'Esta Pastille no tiene servidor GATT.',
		'Diese Pastille hat keinen GATT-Server.'
	],
	'ble.shortInfo': [
		'Pastille sent a device info record shorter than 8 bytes.',
		'La Pastille a envoyé un enregistrement d’informations de moins de 8 octets.',
		'La Pastille envió un registro de información de menos de 8 bytes.',
		'Die Pastille hat einen Geräteinfo-Datensatz von weniger als 8 Byte gesendet.'
	],
	'ble.tapTimeout': [
		'Pastille was not tapped within 60 seconds.',
		'La Pastille n’a pas été touchée en 60 secondes.',
		'No se tocó la Pastille en 60 segundos.',
		'Die Pastille wurde innerhalb von 60 Sekunden nicht angetippt.'
	],
	'ble.noHardwareId': [
		'Pastille did not report its hardware id. Update its firmware, then try again.',
		'La Pastille n’a pas transmis son identifiant matériel. Mettez à jour son firmware, puis réessayez.',
		'La Pastille no informó su id de hardware. Actualiza su firmware e inténtalo de nuevo.',
		'Die Pastille hat ihre Hardware-ID nicht gemeldet. Aktualisieren Sie die Firmware und versuchen Sie es erneut.'
	],
	'ble.unreachable': [
		'Bulle cannot reach your Pastille over Bluetooth.',
		'Bulle ne peut pas joindre votre Pastille en Bluetooth.',
		'Bulle no puede comunicarse con tu Pastille por Bluetooth.',
		'Bulle kann Ihre Pastille nicht per Bluetooth erreichen.'
	],
	'ble.relaying': ['Relaying to Bulle', 'Relais vers Bulle', 'Retransmitiendo a Bulle', 'Weiterleitung an Bulle'],
	'ble.networkError': [
		'The relay to Bulle hit a network error.',
		'Le relais vers Bulle a rencontré une erreur réseau.',
		'La retransmisión a Bulle tuvo un error de red.',
		'Bei der Weiterleitung an Bulle ist ein Netzwerkfehler aufgetreten.'
	],
	'ble.revoked': [
		'This Pastille was revoked, or you signed out.',
		'Cette Pastille a été révoquée, ou vous vous êtes déconnecté.',
		'Esta Pastille fue revocada o cerraste la sesión.',
		'Diese Pastille wurde widerrufen oder Sie haben sich abgemeldet.'
	],
	'ble.closedRetry': [
		'The relay to Bulle closed. Retrying in {seconds} s.',
		'Le relais vers Bulle s’est fermé. Nouvel essai dans {seconds} s.',
		'La retransmisión a Bulle se cerró. Reintentando en {seconds} s.',
		'Die Weiterleitung an Bulle wurde geschlossen. Neuer Versuch in {seconds} s.'
	],
	'ble.disconnected': ['Your Pastille disconnected.', 'Votre Pastille s’est déconnectée.', 'Tu Pastille se desconectó.', 'Ihre Pastille wurde getrennt.'],
	'ble.broke': [
		'The Bluetooth stream from your Pastille broke. Connect it again.',
		'Le flux Bluetooth de votre Pastille s’est interrompu. Reconnectez-la.',
		'Se interrumpió el flujo Bluetooth de tu Pastille. Vuelve a conectarla.',
		'Der Bluetooth-Stream Ihrer Pastille ist abgebrochen. Verbinden Sie sie erneut.'
	],
	'ble.stopped': ['Relay stopped.', 'Relais arrêté.', 'Retransmisión detenida.', 'Weiterleitung beendet.'],
	'ble.encryptedOnly': [
		'Pastille only accepts an encrypted Bluetooth link. Pair it in your system Bluetooth settings, then try again.',
		'La Pastille n’accepte qu’une liaison Bluetooth chiffrée. Appairez-la dans les réglages Bluetooth du système, puis réessayez.',
		'La Pastille solo acepta un enlace Bluetooth cifrado. Empareja en los ajustes Bluetooth del sistema e inténtalo de nuevo.',
		'Die Pastille akzeptiert nur eine verschlüsselte Bluetooth-Verbindung. Koppeln Sie sie in den Bluetooth-Einstellungen des Systems und versuchen Sie es erneut.'
	],
	'ble.noAnswer': [
		'Pastille did not answer {method}.',
		'La Pastille n’a pas répondu à {method}.',
		'La Pastille no respondió a {method}.',
		'Die Pastille hat nicht auf {method} geantwortet.'
	],
	'ble.refused': [
		'Pastille refused the request ({code}).',
		'La Pastille a refusé la requête ({code}).',
		'La Pastille rechazó la solicitud ({code}).',
		'Die Pastille hat die Anfrage abgelehnt ({code}).'
	],
	'login.title': ['Log in - Bulle', 'Connexion - Bulle', 'Iniciar sesión - Bulle', 'Anmelden - Bulle'],
	'login.tagline': [
		'Your assistant. Your suite.',
		'Votre assistant. Votre suite.',
		'Tu asistente. Tu suite.',
		'Ihr Assistent. Ihre Suite.'
	],
	'login.openFailed': [
		'Could not open the sign-in page.',
		'Impossible d’ouvrir la page de connexion.',
		'No se pudo abrir la página de inicio de sesión.',
		'Die Anmeldeseite konnte nicht geöffnet werden.'
	],
	'login.browserHint': [
		'Sign-in opens in your browser, then brings you back here.',
		'La connexion s’ouvre dans votre navigateur, puis vous ramène ici.',
		'El inicio de sesión se abre en tu navegador y luego te trae de vuelta aquí.',
		'Die Anmeldung öffnet sich im Browser und führt Sie danach hierher zurück.'
	],
	'login.copyright': ['© {year} Bulle by Facile.', '© {year} Bulle par Facile.', '© {year} Bulle de Facile.', '© {year} Bulle von Facile.'],
	'login.native.title': ['Sign in to Bulle', 'Se connecter à Bulle', 'Inicia sesión en Bulle', 'Bei Bulle anmelden'],
	'login.native.body': [
		'Use your Facile account. Your sessions, jobs and approvals follow you from the web.',
		'Utilisez votre compte Facile. Vos sessions, tâches et validations vous suivent depuis le web.',
		'Usa tu cuenta Facile. Tus sesiones, tareas y aprobaciones te siguen desde la web.',
		'Nutzen Sie Ihr Facile-Konto. Ihre Sitzungen, Aufgaben und Freigaben folgen Ihnen vom Web.'
	],
	'login.desktop.title': ['Your assistant. On your desktop.', 'Votre assistant. Sur votre bureau.', 'Tu asistente. En tu escritorio.', 'Ihr Assistent. Auf Ihrem Desktop.'],
	'login.desktop.body': [
		'Use your Facile account. Your sessions, jobs and approvals follow you from the web and your phone.',
		'Utilisez votre compte Facile. Vos sessions, tâches et validations vous suivent depuis le web et votre téléphone.',
		'Usa tu cuenta Facile. Tus sesiones, tareas y aprobaciones te siguen desde la web y tu teléfono.',
		'Nutzen Sie Ihr Facile-Konto. Ihre Sitzungen, Aufgaben und Freigaben folgen Ihnen vom Web und von Ihrem Telefon.'
	],
	'login.point.suite.title': ['Works across your suite', 'Fonctionne dans toute votre suite', 'Funciona en toda tu suite', 'Funktioniert in Ihrer ganzen Suite'],
	'login.point.suite.body': [
		'It reads your mail, calendar and tasks in the Facile apps, and answers in seconds.',
		'Il lit vos e-mails, votre agenda et vos tâches dans les applications Facile, et répond en quelques secondes.',
		'Lee tu correo, calendario y tareas en las apps de Facile y responde en segundos.',
		'Es liest Ihre E-Mails, Ihren Kalender und Ihre Aufgaben in den Facile-Apps und antwortet in Sekunden.'
	],
	'login.point.computer.title': ['It has its own computer', 'Il a son propre ordinateur', 'Tiene su propio ordenador', 'Es hat seinen eigenen Computer'],
	'login.point.computer.body': [
		'Code and scripts run on a private machine in the cloud, never on this one.',
		'Le code et les scripts s’exécutent sur une machine privée dans le cloud, jamais sur celle-ci.',
		'El código y los scripts se ejecutan en una máquina privada en la nube, nunca en esta.',
		'Code und Skripte laufen auf einer privaten Cloud-Maschine, nie auf diesem Gerät.'
	],
	'login.point.control.title': ['You stay in control', 'Vous gardez la main', 'Tú mantienes el control', 'Sie behalten die Kontrolle'],
	'login.point.control.body': [
		'Risky actions wait for your approval before they run.',
		'Les actions risquées attendent votre validation avant de s’exécuter.',
		'Las acciones arriesgadas esperan tu aprobación antes de ejecutarse.',
		'Riskante Aktionen warten auf Ihre Freigabe, bevor sie ausgeführt werden.'
	]
} as const satisfies Record<string, Row>;
