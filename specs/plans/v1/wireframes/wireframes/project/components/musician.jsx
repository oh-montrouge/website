/* global React, MOCK, fmtDate, fmtTime, fmtDay, RsvpBadge, EventTypeBadge, SpecNote, Breadcrumb */
const { useState } = React;

// ============================================================
// EVENT LIST — musician view
// ============================================================
function EventListScreen({ navigate, annotations, state = 'populated' }) {
  if (state === 'empty') {
    return (
      <div className="page">
        <div className="page__header">
          <div className="page__title"><h1>Événements</h1><p>Vos répétitions, concerts et autres rendez-vous.</p></div>
        </div>
        <div className="empty">
          <div className="empty__icon">♪</div>
          <h3>Aucun événement à venir</h3>
          <p>Les événements apparaîtront ici dès qu'un administrateur en aura créé. Les événements passés restent visibles pendant 30 jours.</p>
        </div>
        <SpecNote show={annotations}><span>État vide — aucun événement créé ou tous passés depuis plus de 30 jours.</span></SpecNote>
      </div>
    );
  }

  const past = MOCK.events.filter(e => e.past);
  const upcoming = MOCK.events.filter(e => !e.past);

  const renderRow = (ev) => {
    const d = fmtDay(ev.date);
    return (
      <div key={ev.id} className={`event-row ${ev.past ? 'event-row--past' : ''}`} onClick={() => navigate('event-detail', { id: ev.id })}>
        <div className="event-date">
          <div className="event-date__day">{d.day}</div>
          <div className="event-date__month">{d.month}</div>
          <div className="event-date__year">{d.year}</div>
        </div>
        <div className="event-meta">
          <h3>{ev.name}</h3>
          <div className="event-meta__sub">
            <span>🕐 {fmtTime(ev.date)}</span>
            <span>·</span>
            <span>📍 {ev.venue}</span>
          </div>
        </div>
        <EventTypeBadge type={ev.type} />
        <RsvpBadge state={ev.rsvp} />
      </div>
    );
  };

  return (
    <div className="page">
      <div className="page__header">
        <div className="page__title">
          <h1>Événements</h1>
          <p>Répétitions, concerts et autres rendez-vous. Les événements plus vieux que 30 jours ne sont plus affichés.</p>
        </div>
      </div>

      {past.length > 0 && <>
        <div className="section-label">Récent (30 derniers jours)</div>
        <div className="card" style={{ overflow: 'hidden' }}>{past.map(renderRow)}</div>
      </>}
      <div className="section-label">À venir · {upcoming.length} événements</div>
      <div className="card" style={{ overflow: 'hidden' }}>{upcoming.map(renderRow)}</div>

      <SpecNote show={annotations}>
        <span>Affiche les événements des 30 derniers jours + tous ceux à venir, triés par date.</span>
        <span>La pastille RSVP montre votre propre réponse. « Sans réponse » si votre compte existait au moment de la création de l'événement.</span>
        <span>Aucune pastille si l'événement a été créé avant l'activation de votre compte.</span>
      </SpecNote>
    </div>
  );
}

// ============================================================
// PUPITRE TABLE — headcount per instrument (for stage setup)
// ============================================================
function PupitreTable({ rsvpList, eventType, onInstrumentClick, activeInstrument }) {
  // Aggregate by instrument: yes / maybe / no / unanswered
  const rows = {};
  rsvpList.forEach(r => {
    // For concerts, use the "Instrument joué" (may differ from main). For other event types,
    // the RSVP doesn't carry an instrument so we use the musician's own instrument field.
    const inst = r.instrument;
    if (!rows[inst]) rows[inst] = { yes: 0, maybe: 0, no: 0, unanswered: 0 };
    rows[inst][r.state]++;
  });
  const instruments = Object.keys(rows).sort((a, b) => rows[b].yes - rows[a].yes);
  const totals = instruments.reduce((acc, i) => {
    acc.yes += rows[i].yes; acc.maybe += rows[i].maybe; acc.no += rows[i].no; acc.unanswered += rows[i].unanswered;
    return acc;
  }, { yes: 0, maybe: 0, no: 0, unanswered: 0 });

  return (
    <div className="pupitre">
      <div className="pupitre__title">
        <span className="pupitre__icon">🪑</span>
        <div>
          <strong>Effectif par pupitre</strong>
          <span className="muted" style={{ fontSize: 12, marginLeft: 8 }}>— pour l'installation (chaises, pupitres)</span>
        </div>
      </div>
      <table className="table table--compact pupitre__table">
        <thead>
          <tr>
            <th><span className="pupitre__th-full">{eventType === 'concert' ? 'Pupitre (instrument joué)' : 'Pupitre'}</span><span className="pupitre__th-abbr">Pupitre</span></th>
            <th className="pupitre__col-num"><span className="pupitre__th-full">Présents</span><span className="pupitre__th-abbr">Prés.</span></th>
            <th className="pupitre__col-num pupitre__col-muted"><span className="pupitre__th-full">Peut-être</span><span className="pupitre__th-abbr">?</span></th>
            <th className="pupitre__col-num pupitre__col-muted"><span className="pupitre__th-full">Absents</span><span className="pupitre__th-abbr">Abs.</span></th>
            <th className="pupitre__col-num pupitre__col-muted"><span className="pupitre__th-full">Sans rép.</span><span className="pupitre__th-abbr">—</span></th>
          </tr>
        </thead>
        <tbody>
          {instruments.map(inst => {
            const isActive = activeInstrument === inst;
            const isClickable = !!onInstrumentClick;
            return (
              <tr
                key={inst}
                className={`${isClickable ? 'pupitre__row--clickable' : ''} ${isActive ? 'pupitre__row--active' : ''}`}
                onClick={isClickable ? () => onInstrumentClick(isActive ? null : inst) : undefined}
                title={isClickable ? (isActive ? 'Retirer le filtre' : `Filtrer sur ${inst}`) : undefined}
              >
                <td data-label="Pupitre">
                  {inst}
                  {isActive && <span className="pupitre__active-mark"> ✓ filtré</span>}
                </td>
                <td data-label="Présents" className="pupitre__col-num">
                  <strong style={{ color: rows[inst].yes > 0 ? 'var(--ohm-bordeaux)' : 'var(--ink-4)' }}>
                    {rows[inst].yes}
                  </strong>
                </td>
                <td data-label="Peut-être" className="pupitre__col-num pupitre__col-muted">{rows[inst].maybe || '—'}</td>
                <td data-label="Absents" className="pupitre__col-num pupitre__col-muted">{rows[inst].no || '—'}</td>
                <td data-label="Sans rép." className="pupitre__col-num pupitre__col-muted">{rows[inst].unanswered || '—'}</td>
              </tr>
            );
          })}
          <tr className="pupitre__total">
            <td data-label="Total"><strong>Total</strong></td>
            <td data-label="Présents" className="pupitre__col-num"><strong>{totals.yes}</strong></td>
            <td data-label="Peut-être" className="pupitre__col-num">{totals.maybe}</td>
            <td data-label="Absents" className="pupitre__col-num">{totals.no}</td>
            <td data-label="Sans rép." className="pupitre__col-num">{totals.unanswered}</td>
          </tr>
        </tbody>
      </table>
    </div>
  );
}

// ============================================================
// EVENT DETAIL — reacts to event type
// ============================================================
function EventDetailScreen({ navigate, annotations, eventType = 'concert', role = 'musician' }) {
  const [rsvp, setRsvp] = useState('yes');
  const [instrument, setInstrument] = useState('Clarinette');
  const [fieldValues, setFieldValues] = useState({ carpool: 'yes', meal: 'vegetarian', arrival: '' });

  const isAdmin = role === 'admin';

  // Local mutable copy of the RSVP list so admin edits can be reflected live
  const [liveRsvps, setLiveRsvps] = useState(() => MOCK.rsvpList.map(r => ({ ...r })));
  // Track which rows an admin has edited in this session (for "modifié par admin" tag)
  const [editedByAdmin, setEditedByAdmin] = useState(() => new Set());

  // Filter/search state — available to everyone
  const [searchQ, setSearchQ] = useState('');
  const [filterState, setFilterState] = useState('all');     // all | yes | no | maybe | unanswered
  const [filterInstrument, setFilterInstrument] = useState(null); // null or instrument name (concerts only)

  const updateRsvp = (idx, patch) => {
    setLiveRsvps(prev => {
      const next = [...prev];
      next[idx] = { ...next[idx], ...patch };
      return next;
    });
    setEditedByAdmin(prev => {
      const next = new Set(prev);
      next.add(idx);
      return next;
    });
  };

  // Apply filters in order: search → state → instrument (concert only)
  const filtered = liveRsvps
    .map((r, origIdx) => ({ ...r, origIdx }))
    .filter(r => !searchQ || r.name.toLowerCase().includes(searchQ.toLowerCase()))
    .filter(r => filterState === 'all' || r.state === filterState)
    .filter(r => eventType !== 'concert' || !filterInstrument || r.instrument === filterInstrument);

  const hasActiveFilter = searchQ || filterState !== 'all' || filterInstrument;
  const clearFilters = () => { setSearchQ(''); setFilterState('all'); setFilterInstrument(null); };

  const events = {
    concert:   { name: 'Concert de printemps',    date: '2026-05-16T20:30', venue: 'Beffroi de Montrouge', details: 'Programme : Holst (1ère suite), Reed (Armenian Dances), Bernstein (West Side Story). Appel 19h30, tenue noire.' },
    rehearsal: { name: 'Répétition hebdomadaire', date: '2026-04-22T20:00', venue: 'Salle Moulin, Montrouge', details: 'Travail des Armenian Dances (partitions à télécharger sur le Drive). Arrivée 19h45 pour installation.' },
    other:     { name: 'Week-end de travail',     date: '2026-06-07T09:00', venue: 'Centre de loisirs de Champigny', details: 'Deux jours de travail intensif. Merci de confirmer covoiturage et repas avant le 1er juin.' },
  };
  const ev = events[eventType];

  return (
    <div className="page page--wide">
      <Breadcrumb items={[{ label: 'Événements', route: 'events' }, { label: ev.name }]} navigate={navigate} />
      <div className="page__header" style={{ alignItems: 'center' }}>
        <div className="page__title">
          <div style={{ display: 'flex', gap: 10, alignItems: 'center', marginBottom: 8 }}>
            <EventTypeBadge type={eventType} />
            <span className="muted" style={{ fontSize: 14 }}>{fmtDate(ev.date)} · {fmtTime(ev.date)}</span>
          </div>
          <h1>{ev.name}</h1>
          <p>📍 {ev.venue}</p>
        </div>
      </div>

      <div className="split">
        <div className="stack" style={{ '--gap': '24px' }}>
          <div className="card">
            <div className="card__header"><h4>Détails</h4></div>
            <div className="card__body"><p style={{ margin: 0 }}>{ev.details}</p></div>
          </div>

          {/* RSVP LIST */}
          <div className="card">
            <div className="card__header">
              <h4>Réponses des musiciens</h4>
              <span className="muted" style={{ fontSize: 13 }}>
                {liveRsvps.length} musiciens · {liveRsvps.filter(r=>r.state==='yes').length} présents
              </span>
            </div>

            {/* Headcount by pupitre — for stage setup (chairs + stands) */}
            <PupitreTable
              rsvpList={liveRsvps}
              eventType={eventType}
              onInstrumentClick={eventType === 'concert' ? setFilterInstrument : undefined}
              activeInstrument={filterInstrument}
            />

            {/* FILTER BAR — visible to everyone, admin gets inline edit below */}
            <div className="rsvp-filters">
              <div className="rsvp-filters__row">
                <div className="rsvp-filters__search">
                  <span className="rsvp-filters__search-icon">🔍</span>
                  <input
                    type="text"
                    className="input input--bare"
                    placeholder="Rechercher par nom…"
                    value={searchQ}
                    onChange={e => setSearchQ(e.target.value)}
                  />
                  {searchQ && (
                    <button className="rsvp-filters__clear-btn" onClick={() => setSearchQ('')} title="Effacer">✕</button>
                  )}
                </div>
                {eventType === 'concert' && (
                  <select
                    className="select select--compact"
                    value={filterInstrument || ''}
                    onChange={e => setFilterInstrument(e.target.value || null)}
                  >
                    <option value="">Tous les pupitres</option>
                    {MOCK.instruments.map(i => <option key={i} value={i}>{i}</option>)}
                  </select>
                )}
              </div>
              <div className="rsvp-filters__chips">
                {[
                  { v: 'all', label: 'Tous', count: liveRsvps.length },
                  { v: 'yes', label: '✓ Présents', count: liveRsvps.filter(r=>r.state==='yes').length },
                  { v: 'maybe', label: '? Peut-être', count: liveRsvps.filter(r=>r.state==='maybe').length },
                  { v: 'no', label: '✗ Absents', count: liveRsvps.filter(r=>r.state==='no').length },
                  { v: 'unanswered', label: '— Sans réponse', count: liveRsvps.filter(r=>r.state==='unanswered').length, highlight: true },
                ].map(c => (
                  <button
                    key={c.v}
                    className={`chip ${filterState === c.v ? 'chip--active' : ''} ${c.highlight && filterState !== c.v ? 'chip--warn' : ''}`}
                    onClick={() => setFilterState(c.v)}
                  >
                    {c.label}
                    <span className="chip__count">{c.count}</span>
                  </button>
                ))}
                {hasActiveFilter && (
                  <button className="chip chip--link" onClick={clearFilters}>Réinitialiser</button>
                )}
              </div>
              {isAdmin && (
                <div className="rsvp-filters__admin-note">
                  <span className="badge badge--admin" style={{ fontSize: 10 }}>Admin</span>
                  <span style={{ fontSize: 12, color: 'var(--ink-3)' }}>
                    Vous pouvez modifier la réponse d'un musicien en cliquant sur les boutons d'état.
                  </span>
                </div>
              )}
            </div>

            <table className="table rsvp-table">
              <thead>
                <tr>
                  <th>Musicien</th>
                  {eventType === 'concert' && <th>Instrument joué</th>}
                  {eventType === 'other' && <th>Covoiturage</th>}
                  {eventType === 'other' && <th>Repas</th>}
                  <th style={{ width: isAdmin ? 320 : 140 }}>Réponse</th>
                </tr>
              </thead>
              <tbody>
                {filtered.length === 0 && (
                  <tr>
                    <td colSpan={eventType === 'other' ? 4 : eventType === 'concert' ? 3 : 2} style={{ textAlign: 'center', padding: '28px 12px', color: 'var(--ink-3)' }}>
                      Aucun musicien ne correspond à ces filtres.
                    </td>
                  </tr>
                )}
                {filtered.map((r) => {
                  const i = r.origIdx;
                  const adminEdited = editedByAdmin.has(i);
                  return (
                    <tr key={i}>
                      <td data-label="Musicien">
                        <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}>
                          <span>{r.name}</span>
                          {adminEdited && (
                            <span className="tag tag--admin" title="Modifié par un administrateur">modifié par admin</span>
                          )}
                        </div>
                      </td>
                      {eventType === 'concert' && (
                        <td data-label="Instrument">
                          {r.state === 'yes' ? (
                            isAdmin ? (
                              <select
                                className="select select--inline"
                                value={r.instrument || ''}
                                onChange={e => updateRsvp(i, { instrument: e.target.value })}
                              >
                                {MOCK.instruments.map(inst => <option key={inst} value={inst}>{inst}</option>)}
                              </select>
                            ) : (
                              <span className="muted">{r.instrument}</span>
                            )
                          ) : <span className="muted">—</span>}
                        </td>
                      )}
                      {eventType === 'other' && <td data-label="Covoiturage" className="muted">{r.state === 'yes' ? (i % 2 ? 'Propose' : 'Cherche') : '—'}</td>}
                      {eventType === 'other' && <td data-label="Repas" className="muted">{r.state === 'yes' ? (i % 3 === 0 ? 'Végétarien' : 'Standard') : '—'}</td>}
                      <td data-label="Réponse">
                        {isAdmin ? (
                          <div className="rsvp-inline">
                            {[
                              { v: 'yes',        label: '✓', title: 'Présent',    cls: 'yes' },
                              { v: 'maybe',      label: '?', title: 'Peut-être',  cls: 'maybe' },
                              { v: 'no',         label: '✗', title: 'Absent',     cls: 'no' },
                              { v: 'unanswered', label: '—', title: 'Sans réponse', cls: 'unanswered' },
                            ].map(o => (
                              <button
                                key={o.v}
                                className={`rsvp-inline__btn ${r.state === o.v ? `rsvp-inline__btn--active rsvp-inline__btn--${o.cls}` : ''}`}
                                onClick={() => updateRsvp(i, { state: o.v })}
                                title={o.title}
                              >
                                {o.label}
                              </button>
                            ))}
                          </div>
                        ) : (
                          <RsvpBadge state={r.state}/>
                        )}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </div>

        {/* RSVP SIDEBAR */}
        <aside className="card event-rsvp" style={{ position: 'sticky', top: 'calc(var(--header-h) + 16px)' }}>
          <div className="card__header"><h4>Votre réponse</h4></div>
          <div className="card__body stack" style={{ '--gap': '16px' }}>
            <div className="radio-group">
              {[
                { v: 'yes',   label: '✓ Présent',  cls: 'yes' },
                { v: 'no',    label: '✗ Absent',   cls: 'no' },
                { v: 'maybe', label: '? Peut-être', cls: 'maybe' },
              ].map(o => (
                <button key={o.v} className={`radio-pill ${rsvp === o.v ? `radio-pill--active radio-pill--${o.cls}` : ''}`} onClick={() => setRsvp(o.v)}>{o.label}</button>
              ))}
            </div>

            {/* CONCERT — instrument selection when yes */}
            {eventType === 'concert' && rsvp === 'yes' && (
              <div className="field">
                <label className="field__label">Instrument joué <span className="req">*</span></label>
                <select className="select" value={instrument} onChange={e => setInstrument(e.target.value)}>
                  {MOCK.instruments.map(i => <option key={i}>{i}</option>)}
                </select>
                <div className="field__hint">Votre instrument principal est pré-sélectionné. Changez uniquement si vous jouez un autre pupitre pour ce concert.</div>
              </div>
            )}

            {/* OTHER — custom fields when yes */}
            {eventType === 'other' && rsvp === 'yes' && (
              <div className="stack" style={{ '--gap': '14px' }}>
                <div style={{ fontSize: 12, color: 'var(--ink-3)', textTransform: 'uppercase', letterSpacing: '.08em', fontWeight: 600 }}>Informations complémentaires</div>
                <div className="field">
                  <label className="field__label">Covoiturage <span className="req">*</span></label>
                  <select className="select" value={fieldValues.carpool} onChange={e => setFieldValues({...fieldValues, carpool: e.target.value})}>
                    <option value="">— Choisir —</option>
                    <option value="propose">Je propose (voiture disponible)</option>
                    <option value="yes">Je cherche un covoiturage</option>
                    <option value="no">Je n'en ai pas besoin</option>
                  </select>
                </div>
                <div className="field">
                  <label className="field__label">Repas <span className="req">*</span></label>
                  <select className="select" value={fieldValues.meal} onChange={e => setFieldValues({...fieldValues, meal: e.target.value})}>
                    <option value="">— Choisir —</option>
                    <option value="standard">Standard</option>
                    <option value="vegetarian">Végétarien</option>
                    <option value="vegan">Végan</option>
                  </select>
                </div>
                <div className="field">
                  <label className="field__label">Heure d'arrivée (samedi)</label>
                  <input className="input" value={fieldValues.arrival} onChange={e => setFieldValues({...fieldValues, arrival: e.target.value})} placeholder="ex. 9h30"/>
                  <div className="field__hint">Facultatif</div>
                </div>
              </div>
            )}

            <button className="btn btn--primary" style={{ width: '100%' }}>Enregistrer ma réponse</button>
          </div>
        </aside>
      </div>

      <SpecNote show={annotations}>
        {eventType === 'concert' && <span>Concert → si RSVP = oui, sélection d'instrument obligatoire (défaut : instrument principal).</span>}
        {eventType === 'rehearsal' && <span>Répétition → pas de champ supplémentaire, tous les états RSVP sont valides.</span>}
        {eventType === 'other' && <span>Autre → l'admin définit des champs personnalisés. Les champs obligatoires doivent être remplis pour enregistrer un RSVP « présent ».</span>}
        <span>Liste RSVP visible de tous les musiciens authentifiés ; affiche nom + état (+ instrument pour concerts, + réponses aux champs pour « autre »).</span>
        <span>Recherche et filtres disponibles pour tous ; seuls les admins peuvent modifier la réponse d'un autre musicien (édition en ligne).</span>
        {eventType === 'concert' && <span>Le tableau « effectif par pupitre » est cliquable : cliquer un pupitre applique le filtre correspondant à la liste.</span>}
      </SpecNote>
    </div>
  );
}

// ============================================================
// PROFILE — musician's own profile (read-only)
// ============================================================
function ProfileScreen({ annotations }) {
  return (
    <div className="page page--narrow">
      <div className="page__header">
        <div className="page__title">
          <h1>Mon profil</h1>
          <p>Vos informations personnelles. Contactez un administrateur pour toute modification.</p>
        </div>
      </div>

      <div className="row" style={{ marginBottom: 32, gap: 20 }}>
        <span className="avatar avatar--lg">MD</span>
        <div>
          <h2 style={{ fontSize: 24 }}>Marie Dubois</h2>
          <div className="muted">Clarinette · Membre depuis septembre 2018</div>
        </div>
      </div>

      <div className="card">
        <div className="card__body">
          <dl className="kv">
            <dt>Prénom, nom</dt><dd>Marie Dubois</dd>
            <dt>Adresse e-mail</dt><dd>marie.dubois@ohm-montrouge.fr</dd>
            <dt>Instrument principal</dt><dd>Clarinette</dd>
            <dt>Date de naissance</dt><dd>14 mars 1987</dd>
            <dt>Téléphone</dt><dd>06 12 34 56 78</dd>
            <dt>Adresse</dt><dd>42 rue de la Paix, 92120 Montrouge</dd>
          </dl>
        </div>
        <div className="card__footer" style={{ background: 'var(--ohm-gold-50)', borderTop: '1px solid var(--ohm-gold-100)' }}>
          <div style={{ fontSize: 13, color: 'var(--ohm-gold-ink)' }}>
            ℹ Pour retirer votre consentement concernant téléphone et adresse, contactez un administrateur.
          </div>
        </div>
      </div>

      <SpecNote show={annotations}>
        <span>Profil en lecture seule (V1). Affiche téléphone/adresse uniquement si le consentement est donné.</span>
        <span>Notice statique sur le retrait du consentement — toujours visible.</span>
      </SpecNote>
    </div>
  );
}

Object.assign(window, { EventListScreen, EventDetailScreen, ProfileScreen, PupitreTable });
