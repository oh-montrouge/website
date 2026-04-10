# OHM Website — Homepage

> **Depends on:** [Information Model](00-information-model.md)

---

## Purpose

The homepage is the public landing page of the application. It briefly introduces the
Orchestre d'Harmonie de Montrouge to any visitor — member or not.

---

## Access

The homepage is public: no authentication required. Authenticated users and unauthenticated
visitors see the same page. There is no automatic redirect away from `/` for authenticated
users; they navigate to other sections via the normal navigation.

---

## Content

The page contains a short presentation of the association: its name, a brief description,
and optionally an image or logo. The exact copy is provided by the association.

Content is static: it is bundled with the application in a template. There is no admin UI
to edit it in V1; changes require a new deployment.

---

## Navigation

The page includes navigation appropriate to the visitor's authentication state:

- **Unauthenticated:** a link to the login page (`/connexion`) is visible.
- **Authenticated:** the standard application navigation is shown (events, profile, etc.).

The privacy notice link (persistent in the footer for authenticated users, per
[Privacy and Consent](06-privacy-and-consent.md)) is also present on the homepage.
