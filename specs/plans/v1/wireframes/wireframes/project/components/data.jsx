/* global React */
const { useState } = React;

// Mock data shared across all screens
const MOCK = {
  events: [
    { id: 'e1', name: 'Répétition hebdomadaire', type: 'rehearsal', date: '2026-04-22T20:00', venue: 'Salle Moulin, Montrouge', past: false, rsvp: 'yes' },
    { id: 'e2', name: 'Concert de printemps', type: 'concert', date: '2026-05-16T20:30', venue: 'Beffroi de Montrouge', past: false, rsvp: 'yes', instrument: 'Clarinette' },
    { id: 'e3', name: 'Week-end de travail', type: 'other', date: '2026-06-07T09:00', venue: 'Centre de loisirs de Champigny', past: false, rsvp: 'maybe' },
    { id: 'e4', name: 'Aubade du 14 juillet', type: 'concert', date: '2026-07-14T11:00', venue: 'Parc Messier', past: false, rsvp: 'unanswered' },
    { id: 'e5', name: 'Répétition de rentrée', type: 'rehearsal', date: '2026-04-08T20:00', venue: 'Salle Moulin, Montrouge', past: true, rsvp: 'yes' },
  ],
  musicians: [
    { id: 'm1', first: 'Marie',      last: 'Dubois',    instrument: 'Clarinette',   status: 'active',     admin: true,  email: 'marie.dubois@ohm-montrouge.fr', inscription: '2018-09' },
    { id: 'm2', first: 'Antoine',    last: 'Leroux',    instrument: 'Trompette',    status: 'active',     admin: false, email: 'a.leroux@example.com',          inscription: '2021-09' },
    { id: 'm3', first: 'Camille',    last: 'Petit',     instrument: 'Flûte',        status: 'active',     admin: true,  email: 'camille.petit@example.com',     inscription: '2019-09' },
    { id: 'm4', first: 'Hugo',       last: 'Bernard',   instrument: 'Saxophone alto', status: 'pending', admin: false, email: 'hugo.bernard@example.com',      inscription: null },
    { id: 'm5', first: 'Léa',        last: 'Moreau',    instrument: 'Hautbois',     status: 'active',     admin: false, email: 'lea.moreau@example.com',        inscription: '2023-09' },
    { id: 'm6', first: 'Julien',     last: 'Martin',    instrument: 'Percussions',  status: 'active',     admin: false, email: 'j.martin@example.com',          inscription: '2015-09' },
    { id: 'm7', first: '—',          last: '—',         instrument: 'Cor',          status: 'anonymized', admin: false, email: null,                            inscription: '2012-09' },
    { id: 'm8', first: 'Sophie',     last: 'Garcia',    instrument: 'Basson',       status: 'active',     admin: false, email: 'sophie.garcia@example.com',     inscription: '2020-09' },
  ],
  seasons: [
    { id: 's1', label: '2025-2026', start: '2025-09-01', end: '2026-08-31', current: true,  payments: 42 },
    { id: 's2', label: '2024-2025', start: '2024-09-01', end: '2025-08-31', current: false, payments: 48 },
    { id: 's3', label: '2023-2024', start: '2023-09-01', end: '2024-08-31', current: false, payments: 45 },
    { id: 's4', label: '2022-2023', start: '2022-09-01', end: '2023-08-31', current: false, payments: 43 },
  ],
  payments: [
    { id: 'p1', season: '2025-2026', amount: 180, date: '2025-09-18', type: 'virement bancaire', comment: '' },
    { id: 'p2', season: '2024-2025', amount: 180, date: '2024-09-22', type: 'chèque', comment: 'Chèque n° 2145' },
    { id: 'p3', season: '2023-2024', amount: 170, date: '2023-10-05', type: 'espèces', comment: '' },
    { id: 'p4', season: '2022-2023', amount: 170, date: '2022-09-14', type: 'chèque', comment: '' },
  ],
  rsvpList: [
    { name: 'Marie Dubois',    instrument: 'Clarinette',   state: 'yes' },
    { name: 'Antoine Leroux',  instrument: 'Trompette',    state: 'yes' },
    { name: 'Camille Petit',   instrument: 'Flûte',        state: 'yes' },
    { name: 'Léa Moreau',      instrument: 'Hautbois',     state: 'yes' },
    { name: 'Julien Martin',   instrument: 'Percussions',  state: 'yes' },
    { name: 'Nicolas Girard',  instrument: 'Clarinette',   state: 'yes' },
    { name: 'Élodie Rey',      instrument: 'Flûte',        state: 'yes' },
    { name: 'Mathieu Blanc',   instrument: 'Trompette',    state: 'yes' },
    { name: 'Sophie Garcia',   instrument: 'Basson',       state: 'maybe' },
    { name: 'Paul Durand',     instrument: 'Cor',          state: 'no' },
    { name: 'Claire Fontaine', instrument: 'Trombone',     state: 'unanswered' },
    { name: 'Thomas Roux',     instrument: 'Tuba',         state: 'unanswered' },
  ],
  instruments: ['Flûte','Hautbois','Clarinette','Basson','Saxophone alto','Saxophone ténor','Cor','Trompette','Trombone','Tuba','Percussions','Chef d\'orchestre'],
};

// Format date for display
function fmtDate(iso) {
  const d = new Date(iso);
  return d.toLocaleDateString('fr-FR', { day: 'numeric', month: 'long', year: 'numeric' });
}
function fmtTime(iso) {
  const d = new Date(iso);
  return d.toLocaleTimeString('fr-FR', { hour: '2-digit', minute: '2-digit' });
}
function fmtDay(iso) {
  const d = new Date(iso);
  return { day: d.getDate(), month: d.toLocaleDateString('fr-FR', { month: 'short' }).replace('.', ''), year: d.getFullYear() };
}

Object.assign(window, { MOCK, fmtDate, fmtTime, fmtDay });
