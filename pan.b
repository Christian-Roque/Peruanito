	switch (t->back) {
	default: Uerror("bad return move");
	case  0: goto R999; /* nothing to undo */

		 /* PROC :init: */

	case 3: // STATE 1
		;
		;
		delproc(0, now._nr_pr-1);
		;
		goto R999;

	case 4: // STATE 2
		;
		;
		delproc(0, now._nr_pr-1);
		;
		goto R999;

	case 5: // STATE 3
		;
		;
		delproc(0, now._nr_pr-1);
		;
		goto R999;

	case 6: // STATE 4
		;
		;
		delproc(0, now._nr_pr-1);
		;
		goto R999;

	case 7: // STATE 6
		;
		p_restor(II);
		;
		;
		goto R999;

		 /* PROC Consumidor */

	case 8: // STATE 1
		;
		XX = 1;
		unrecv(now.procesado, XX-1, 0, ((int)((P2 *)_this)->r), 1);
		((P2 *)_this)->r = trpt->bup.oval;
		;
		;
		goto R999;

	case 9: // STATE 3
		;
		((P2 *)_this)->workers_finalizados = trpt->bup.ovals[1];
	/* 0 */	((P2 *)_this)->r = trpt->bup.ovals[0];
		;
		;
		ungrab_ints(trpt->bup.ovals, 2);
		goto R999;

	case 10: // STATE 4
		;
	/* 0 */	((P2 *)_this)->workers_finalizados = trpt->bup.oval;
		;
		;
		goto R999;

	case 11: // STATE 11
		;
		now.total_consumidos = trpt->bup.oval;
		;
		goto R999;
;
		;
		
	case 13: // STATE 18
		;
		p_restor(II);
		;
		;
		goto R999;

		 /* PROC Worker */

	case 14: // STATE 1
		;
		XX = 1;
		unrecv(now.tareas, XX-1, 0, ((int)((P1 *)_this)->t), 1);
		((P1 *)_this)->t = trpt->bup.oval;
		;
		;
		goto R999;

	case 15: // STATE 2
		;
	/* 0 */	((P1 *)_this)->t = trpt->bup.oval;
		;
		;
		goto R999;

	case 16: // STATE 3
		;
		_m = unsend(now.procesado);
		;
		goto R999;

	case 17: // STATE 7
		;
		now.mutex = trpt->bup.oval;
		;
		goto R999;

	case 18: // STATE 10
		;
		now.en_critica = trpt->bup.oval;
		;
		goto R999;
;
		;
		
	case 20: // STATE 12
		;
		total_unicos = trpt->bup.oval;
		;
		goto R999;

	case 21: // STATE 13
		;
		now.en_critica = trpt->bup.oval;
		;
		goto R999;

	case 22: // STATE 14
		;
		now.mutex = trpt->bup.oval;
		;
		goto R999;

	case 23: // STATE 16
		;
		_m = unsend(now.procesado);
		;
		goto R999;

	case 24: // STATE 22
		;
		p_restor(II);
		;
		;
		goto R999;

		 /* PROC Productor */
;
		;
		
	case 26: // STATE 2
		;
		_m = unsend(now.tareas);
		;
		goto R999;

	case 27: // STATE 3
		;
		((P0 *)_this)->i = trpt->bup.oval;
		;
		goto R999;

	case 28: // STATE 9
		;
		((P0 *)_this)->i = trpt->bup.oval;
		;
		goto R999;
;
		;
		
	case 30: // STATE 11
		;
		_m = unsend(now.tareas);
		;
		goto R999;

	case 31: // STATE 12
		;
		((P0 *)_this)->i = trpt->bup.oval;
		;
		goto R999;

	case 32: // STATE 18
		;
		p_restor(II);
		;
		;
		goto R999;
	}

