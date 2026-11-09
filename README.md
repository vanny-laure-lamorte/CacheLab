# CacheLab

## Sommaire

1. [Présentation du projet](#présentation-du-projet)
2. [Objectifs](#objectifs)
3. [Technologies utilisées](#technologies-utilisées)
4. [Architecture du projet](#architecture-du-projet)
5. [Installation et prérequis](#installation-et-prérequis)
6. [Configuration](#configuration)
7. [Lancement du serveur](#lancement-du-serveur)
8. [Documentation de l'API REST](#documentation-de-lapi-rest)
9. [Choix techniques et structure de données](#choix-techniques-et-structure-de-données)
10. [Sécurité et gestion des erreurs](#sécurité-et-gestion-des-erreurs)
11. [Tests](#tests)
12. [Performances et gestion de la mémoire](#performances-et-gestion-de-la-mémoire)
13. [Limites et améliorations possibles](#limites-et-améliorations-possibles)
14. [Documentation du projet](#documentation-du-projet)
15. [Équipe et planification](#équipe-et-planification)
16. [Licence](#licence)

---

## Présentation du projet

CacheLab est un projet de développement d'un système de cache clé/valeur en mémoire, inspiré des principes de Redis. Les applications web, notamment les sites e-commerce, peuvent subir des ralentissements lorsque le nombre de requêtes augmente fortement. Un cache permet de conserver temporairement en mémoire les données fréquemment consultées afin de limiter les accès à la base de données principale.

Le projet CacheLab propose une solution simple permettant de stocker, récupérer, modifier et supprimer des données à travers une API REST.

Nous avons décidé de développé en **Go** et utilise une structure de données en mémoire pour permettre un accès rapide aux valeurs.

## Objectifs

Les objectifs principaux de CacheLab sont :

- Développer un système de cache clé/valeur fonctionnel
- Implémenter les opérations CRUD : création, lecture, modification et suppression
- Concevoir une API REST permettant aux applications clientes d'interagir avec le cache
- Utiliser une structure de données adaptée à la recherche rapide
- Assurer la sécurité des accès concurrents aux données
- Étudier les moyens de limiter et d'optimiser la consommation mémoire
- Mettre en place des tests unitaires et des tests d'intégration
- Étudier les solutions existantes, notamment Redis, afin de justifier les choix techniques
- Documenter les aspects fonctionnels, techniques, sécuritaires et écoresponsables du projet

Le périmètre initial correspond à un *Minimum Viable Product*, une première version fonctionnelle contenant les fonctionnalités essentielles.

## Technologies utilisées

| Technologie | Utilisation |
|---|---|
| Go | Développement du serveur et de la logique du cache |
| `net/http` | Création du serveur HTTP et gestion des routes REST |
| `encoding/json` | Lecture et génération des données JSON |
| `map[string]string` | Stockage des paires clé/valeur en mémoire |
| `sync.RWMutex` | Protection contre les accès concurrents au stockage |
| `testing` | Écriture et exécution des tests unitaires |
| Git | Gestion des versions du code |
| GitHub | Hébergement du dépôt et collaboration |

Go a été choisi pour ses performances, sa gestion de la concurrence et sa bibliothèque standard adaptée au développement d'un serveur HTTP.

## Architecture du projet

Le projet est organisé en plusieurs parties afin de séparer la logique métier du cache de la gestion des requêtes HTTP.

```text
CacheLab/
├── docs/
├── internal/
│   ├── api/
│   │   └── handlers.go
│   └── cache/
│       ├── store.go
│       └── store_test.go
├── go.mod
├── main.go
└── README.md
```

### Rôle des fichiers

- **`main.go`** : initialise les composants et démarre le serveur HTTP.
- **`internal/api/handlers.go`** : traite les requêtes HTTP, appelle les opérations du cache et construit les réponses.
- **`internal/cache/store.go`** : contient la logique de stockage et les opérations sur les clés et les valeurs.
- **`internal/cache/store_test.go`** : regroupe les tests unitaires du stockage.
- **`go.mod`** : définit le module Go et ses dépendances.
- **`docs/`** : accueille les documents de conception, le rapport de veille et le planning.
- **`README.md`** : présente le projet et explique comment l'installer et l'utiliser.

Cette architecture modulaire facilite la maintenance, les tests et l'évolution du projet.

## Installation et prérequis

### Prérequis

Pour utiliser le projet, il faut disposer de :

- Go version 1.22 ou ultérieure.
- Git pour récupérer le dépôt.
- Un terminal ou un éditeur de code, par exemple Visual Studio Code.

Téléchargement de Go : https://go.dev/dl/

### Cloner le dépôt

Remplacer `URL_DU_DEPOT` par l'adresse réelle du dépôt GitHub.

```bash
git clone URL_DU_DEPOT
cd CacheLab
```

### Préparer le projet

À la racine du dépôt, exécuter :

```bash
go mod tidy
```

Cette commande prépare les dépendances du module Go.

## Configuration

Le serveur utilise la variable d'environnement `PORT` pour définir son port d'écoute.

Si aucune valeur n'est définie, le port `8080` est utilisé par défaut.

Sous Linux ou macOS :

```bash
export PORT=8081
```

Sous Windows PowerShell :

```powershell
$env:PORT="8081"
```

La configuration peut être adaptée au fur et à mesure de l'évolution du projet.

## Lancement du serveur

Depuis le répertoire racine du projet, exécuter :

```bash
go run .
```

Le serveur démarre sur le port configuré, ou sur le port `8080` par défaut.

Pour compiler le projet :

```bash
go build ./...
```

## Documentation de l'API REST

L'API REST permet de manipuler les paires clé/valeur du cache. Les données échangées sont au format JSON.

### Endpoints disponibles

| Méthode | Endpoint | Description |
|---|---|---|
| `POST` | `/keys` | Créer une nouvelle paire clé/valeur |
| `GET` | `/keys/{key}` | Récupérer la valeur d'une clé |
| `PUT` | `/keys/{key}` | Modifier la valeur d'une clé existante |
| `DELETE` | `/keys/{key}` | Supprimer une clé |
| `GET` | `/keys` | Lister les clés stockées (fonctionnalité complémentaire) |

### Créer une clé — POST /keys

Exemple de requête :

```http
POST /keys
Content-Type: application/json
```

Corps JSON :

```json
{
  "key": "produit:42",
  "value": "clavier"
}
```

Exemple avec `curl` :

```bash
curl -i -X POST http://localhost:8080/keys \
  -H "Content-Type: application/json" \
  -d '{"key":"produit:42","value":"clavier"}'
```

La réponse attendue en cas de création réussie utilise le statut HTTP `201 Created`.

### Lire une clé — GET /keys/{key}

Exemple :

```bash
curl -i http://localhost:8080/keys/produit:42
```

En cas de succès, l'API retourne la clé et sa valeur au format JSON.

### Modifier une clé — PUT /keys/{key}

Exemple :

```bash
curl -i -X PUT http://localhost:8080/keys/produit:42 \
  -H "Content-Type: application/json" \
  -d '{"value":"souris"}'
```

La modification concerne une clé existante. Si la clé n'existe pas, l'API doit retourner une erreur adaptée.

### Supprimer une clé — DELETE /keys/{key}

Exemple :

```bash
curl -i -X DELETE http://localhost:8080/keys/produit:42
```

En cas de suppression réussie, le statut HTTP `204 No Content` peut être retourné.

### Lister les clés — GET /keys

Exemple :

```bash
curl -i http://localhost:8080/keys
```

Cette route permet de récupérer la liste des clés présentes dans le cache. Elle constitue une fonctionnalité complémentaire au CRUD obligatoire.

### Codes HTTP

| Code HTTP | Signification |
|---|---|
| `200 OK` | Lecture ou modification réussie |
| `201 Created` | Clé créée |
| `204 No Content` | Clé supprimée |
| `400 Bad Request` | Requête invalide |
| `404 Not Found` | Clé inexistante |
| `409 Conflict` | Clé déjà existante |

Les réponses exactes doivent être vérifiées avec l'implémentation actuelle de l'API.

## Choix techniques et structure de données

### Choix du langage

Go a été retenu pour plusieurs raisons :

- **Performances** : compilation en code natif et bonne efficacité pour les services backend.
- **Concurrence** : primitives intégrées pour gérer les accès simultanés.
- **Simplicité** : syntaxe relativement concise et outils de développement intégrés.
- **Écosystème** : bibliothèque standard permettant de créer une API HTTP sans framework externe obligatoire.

### Structure de données

Le cache utilise une `map[string]string` de Go pour associer chaque clé à une valeur.

Une table de hachage permet généralement d'effectuer une recherche, une insertion ou une suppression en temps moyen O(1). Les performances exactes dépendent notamment de la taille des données et de l'implémentation.

Ce choix est adapté à un cache clé/valeur, car il favorise les accès directs par clé.

### Gestion de la concurrence

Le stockage utilise `sync.RWMutex` pour protéger les accès concurrents :

- `RLock()` autorise plusieurs lectures simultanées.
- `Lock()` protège les opérations de modification.

Cette protection évite les accès concurrents non synchronisés à la map.

### Persistance

Les données sont conservées uniquement en mémoire vive (RAM). Elles sont perdues lorsque le processus s'arrête ou redémarre.

La persistance sur disque n'est pas incluse dans le périmètre initial du cache.

## Sécurité et gestion des erreurs

La sécurité fait partie des points à traiter pour rendre le service utilisable dans un environnement réel.

Les mesures à prévoir comprennent :

- Valider les clés et les valeurs reçues.
- Vérifier le format des corps JSON.
- Limiter la taille des requêtes pour éviter une consommation excessive de mémoire.
- Retourner des codes HTTP cohérents.
- Éviter de renvoyer des informations internes sensibles dans les messages d'erreur.
- Étudier l'authentification si l'API est accessible à des clients externes.
- Limiter l'exposition réseau et sécuriser le déploiement.

Si des données personnelles sont mises en cache, leur traitement doit également prendre en compte les exigences applicables du RGPD, notamment la limitation de conservation et la protection des données.

## Tests

Les tests sont essentiels pour vérifier la fiabilité des opérations du cache.

Exécuter les tests unitaires :

```bash
go test ./...
```

Vérifier la compilation :

```bash
go build ./...
```

Les tests doivent notamment couvrir :

- La création d'une clé.
- La lecture d'une clé existante.
- La modification d'une valeur.
- La suppression d'une clé.
- La lecture d'une clé inexistante.
- La tentative de création d'une clé déjà existante.
- La validation des entrées et les réponses HTTP.
- Les accès concurrents au stockage.

Les résultats doivent être mis à jour au fur et à mesure de l'ajout des tests.

## Performances et gestion de la mémoire

Le cache est conçu pour fournir un accès rapide aux données fréquemment consultées.

Le stockage en RAM évite les accès disque pour les opérations courantes, mais il implique plusieurs contraintes :

- La capacité du cache dépend de la mémoire disponible.
- Les données occupent de la mémoire tant qu'elles restent stockées.
- Les données sont perdues à l'arrêt du processus.
- Une charge élevée peut augmenter la consommation de mémoire et de ressources CPU.

Les pistes d'amélioration comprennent la mesure du temps de réponse, les tests de charge, la limitation de la taille des données et la mise en place d'une politique d'éviction.

L'impact de la consommation de RAM doit également être étudié dans une démarche d'éco-conception.

## Limites et améliorations possibles

La première version se concentre sur le stockage clé/valeur et les opérations CRUD.

Les évolutions possibles sont :

- Ajouter un TTL (*Time To Live*) pour définir la durée de vie des clés.
- Implémenter une politique d'éviction LRU (*Least Recently Used*).
- Définir une limite de mémoire configurable.
- Ajouter des métriques et un système de monitoring.
- Renforcer la validation des entrées.
- Ajouter des tests d'intégration et de charge.
- Étudier des mécanismes de persistance si les besoins le justifient.
- Améliorer la documentation et les procédures de déploiement.

Ces fonctionnalités seront priorisées selon le temps disponible et les besoins du MVP.

## Documentation du projet

Le dossier `docs/` est destiné à accueillir les livrables de conception :

- **Cahier des charges fonctionnel** : besoins, cas d'usage, fonctionnalités prioritaires et contraintes.
- **Cahier des charges technique** : architecture, choix du langage, structure de données, API et sécurité.
- **Rapport de veille** : bases clé/valeur, comparaison avec SQL, analyse de Redis, performances, sécurité et éco-conception.
- **Note de cadrage et planning** : tâches, répartition des rôles, délais et risques.

Ces documents complètent le code source et justifient les choix effectués pendant le développement.

## Équipe et planification

Le projet est réalisé dans le cadre d'un développement en équipe, avec un délai cible de deux semaines pour livrer le MVP.

La planification comprend :

1. Recherche documentaire et analyse des besoins.
2. Définition de l'architecture et des choix techniques.
3. Implémentation du stockage en mémoire et de l'API REST.
4. Tests, corrections et vérification des performances.
5. Finalisation de la documentation et livraison.

La répartition des rôles, les échéances détaillées et les risques sont à préciser dans la note de cadrage.
