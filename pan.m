#define rand	pan_rand
#define pthread_equal(a,b)	((a)==(b))
#if defined(HAS_CODE) && defined(VERBOSE)
	#ifdef BFS_PAR
		bfs_printf("Pr: %d Tr: %d\n", II, t->forw);
	#else
		cpu_printf("Pr: %d Tr: %d\n", II, t->forw);
	#endif
#endif
	switch (t->forw) {
	default: Uerror("bad forward move");
	case 0:	/* if without executable clauses */
		continue;
	case 1: /* generic 'goto' or 'skip' */
		IfNotBlocked
		_m = 3; goto P999;
	case 2: /* generic 'else' */
		IfNotBlocked
		if (trpt->o_pm&1) continue;
		_m = 3; goto P999;

		 /* PROC :init: */
	case 3: // STATE 1 - modelo_verificacion_spin.pml:107 - [(run Productor())] (0:0:0 - 1)
		IfNotBlocked
		reached[3][1] = 1;
		if (!(addproc(II, 1, 0, 0)))
			continue;
		_m = 3; goto P999; /* 0 */
	case 4: // STATE 2 - modelo_verificacion_spin.pml:108 - [(run Worker(0))] (0:0:0 - 1)
		IfNotBlocked
		reached[3][2] = 1;
		if (!(addproc(II, 1, 1, 0)))
			continue;
		_m = 3; goto P999; /* 0 */
	case 5: // STATE 3 - modelo_verificacion_spin.pml:109 - [(run Worker(1))] (0:0:0 - 1)
		IfNotBlocked
		reached[3][3] = 1;
		if (!(addproc(II, 1, 1, 1)))
			continue;
		_m = 3; goto P999; /* 0 */
	case 6: // STATE 4 - modelo_verificacion_spin.pml:110 - [(run Consumidor())] (0:0:0 - 1)
		IfNotBlocked
		reached[3][4] = 1;
		if (!(addproc(II, 1, 2, 0)))
			continue;
		_m = 3; goto P999; /* 0 */
	case 7: // STATE 6 - modelo_verificacion_spin.pml:112 - [-end-] (0:0:0 - 1)
		IfNotBlocked
		reached[3][6] = 1;
		if (!delproc(1, II)) continue;
		_m = 3; goto P999; /* 0 */

		 /* PROC Consumidor */
	case 8: // STATE 1 - modelo_verificacion_spin.pml:85 - [procesado?r] (0:0:1 - 1)
		reached[2][1] = 1;
		if (q_len(now.procesado) == 0) continue;

		XX=1;
		(trpt+1)->bup.oval = ((int)((P2 *)_this)->r);
		;
		((P2 *)_this)->r = qrecv(now.procesado, XX-1, 0, 1);
#ifdef VAR_RANGES
		logval("Consumidor:r", ((int)((P2 *)_this)->r));
#endif
		;
		
#ifdef HAS_CODE
		if (readtrail && gui) {
			char simtmp[32];
			sprintf(simvals, "%d?", now.procesado);
		sprintf(simtmp, "%d", ((int)((P2 *)_this)->r)); strcat(simvals, simtmp);		}
#endif
		;
		_m = 4; goto P999; /* 0 */
	case 9: // STATE 2 - modelo_verificacion_spin.pml:87 - [((r==255))] (8:0:2 - 1)
		IfNotBlocked
		reached[2][2] = 1;
		if (!((((int)((P2 *)_this)->r)==255)))
			continue;
		if (TstOnly) return 1; /* TT */
		/* dead 1: r */  (trpt+1)->bup.ovals = grab_ints(2);
		(trpt+1)->bup.ovals[0] = ((P2 *)_this)->r;
#ifdef HAS_CODE
		if (!readtrail)
#endif
			((P2 *)_this)->r = 0;
		/* merge: workers_finalizados = (workers_finalizados+1)(0, 3, 8) */
		reached[2][3] = 1;
		(trpt+1)->bup.ovals[1] = ((int)((P2 *)_this)->workers_finalizados);
		((P2 *)_this)->workers_finalizados = (((int)((P2 *)_this)->workers_finalizados)+1);
#ifdef VAR_RANGES
		logval("Consumidor:workers_finalizados", ((int)((P2 *)_this)->workers_finalizados));
#endif
		;
		_m = 3; goto P999; /* 1 */
	case 10: // STATE 4 - modelo_verificacion_spin.pml:91 - [((workers_finalizados==2))] (0:0:1 - 1)
		IfNotBlocked
		reached[2][4] = 1;
		if (!((((int)((P2 *)_this)->workers_finalizados)==2)))
			continue;
		if (TstOnly) return 1; /* TT */
		/* dead 1: workers_finalizados */  (trpt+1)->bup.oval = ((P2 *)_this)->workers_finalizados;
#ifdef HAS_CODE
		if (!readtrail)
#endif
			((P2 *)_this)->workers_finalizados = 0;
		_m = 3; goto P999; /* 0 */
	case 11: // STATE 11 - modelo_verificacion_spin.pml:98 - [total_consumidos = (total_consumidos+1)] (0:0:1 - 1)
		IfNotBlocked
		reached[2][11] = 1;
		(trpt+1)->bup.oval = ((int)now.total_consumidos);
		now.total_consumidos = (((int)now.total_consumidos)+1);
#ifdef VAR_RANGES
		logval("total_consumidos", ((int)now.total_consumidos));
#endif
		;
		_m = 3; goto P999; /* 0 */
	case 12: // STATE 17 - modelo_verificacion_spin.pml:102 - [assert((total_consumidos==4))] (0:0:0 - 3)
		IfNotBlocked
		reached[2][17] = 1;
		spin_assert((((int)now.total_consumidos)==4), "(total_consumidos==4)", II, tt, t);
		_m = 3; goto P999; /* 0 */
	case 13: // STATE 18 - modelo_verificacion_spin.pml:103 - [-end-] (0:0:0 - 1)
		IfNotBlocked
		reached[2][18] = 1;
		if (!delproc(1, II)) continue;
		_m = 3; goto P999; /* 0 */

		 /* PROC Worker */
	case 14: // STATE 1 - modelo_verificacion_spin.pml:53 - [tareas?t] (0:0:1 - 1)
		reached[1][1] = 1;
		if (q_len(now.tareas) == 0) continue;

		XX=1;
		(trpt+1)->bup.oval = ((int)((P1 *)_this)->t);
		;
		((P1 *)_this)->t = qrecv(now.tareas, XX-1, 0, 1);
#ifdef VAR_RANGES
		logval("Worker:t", ((int)((P1 *)_this)->t));
#endif
		;
		
#ifdef HAS_CODE
		if (readtrail && gui) {
			char simtmp[32];
			sprintf(simvals, "%d?", now.tareas);
		sprintf(simtmp, "%d", ((int)((P1 *)_this)->t)); strcat(simvals, simtmp);		}
#endif
		;
		_m = 4; goto P999; /* 0 */
	case 15: // STATE 2 - modelo_verificacion_spin.pml:55 - [((t==255))] (0:0:1 - 1)
		IfNotBlocked
		reached[1][2] = 1;
		if (!((((int)((P1 *)_this)->t)==255)))
			continue;
		if (TstOnly) return 1; /* TT */
		/* dead 1: t */  (trpt+1)->bup.oval = ((P1 *)_this)->t;
#ifdef HAS_CODE
		if (!readtrail)
#endif
			((P1 *)_this)->t = 0;
		_m = 3; goto P999; /* 0 */
	case 16: // STATE 3 - modelo_verificacion_spin.pml:56 - [procesado!255] (0:0:0 - 1)
		IfNotBlocked
		reached[1][3] = 1;
		if (q_full(now.procesado))
			continue;
#ifdef HAS_CODE
		if (readtrail && gui) {
			char simtmp[64];
			sprintf(simvals, "%d!", now.procesado);
		sprintf(simtmp, "%d", 255); strcat(simvals, simtmp);		}
#endif
		
		qsend(now.procesado, 0, 255, 1);
		_m = 2; goto P999; /* 0 */
	case 17: // STATE 6 - modelo_verificacion_spin.pml:20 - [((mutex==0))] (10:0:1 - 1)
		IfNotBlocked
		reached[1][6] = 1;
		if (!((((int)now.mutex)==0)))
			continue;
		/* merge: mutex = 1(0, 7, 10) */
		reached[1][7] = 1;
		(trpt+1)->bup.oval = ((int)now.mutex);
		now.mutex = 1;
#ifdef VAR_RANGES
		logval("mutex", ((int)now.mutex));
#endif
		;
		_m = 3; goto P999; /* 1 */
	case 18: // STATE 10 - modelo_verificacion_spin.pml:66 - [en_critica = (en_critica+1)] (0:0:1 - 1)
		IfNotBlocked
		reached[1][10] = 1;
		(trpt+1)->bup.oval = ((int)now.en_critica);
		now.en_critica = (((int)now.en_critica)+1);
#ifdef VAR_RANGES
		logval("en_critica", ((int)now.en_critica));
#endif
		;
		_m = 3; goto P999; /* 0 */
	case 19: // STATE 11 - modelo_verificacion_spin.pml:67 - [assert((en_critica==1))] (0:0:0 - 1)
		IfNotBlocked
		reached[1][11] = 1;
		spin_assert((((int)now.en_critica)==1), "(en_critica==1)", II, tt, t);
		_m = 3; goto P999; /* 0 */
	case 20: // STATE 12 - modelo_verificacion_spin.pml:69 - [total_unicos = (total_unicos+1)] (0:0:1 - 1)
		IfNotBlocked
		reached[1][12] = 1;
		(trpt+1)->bup.oval = ((int)total_unicos);
		total_unicos = (((int)total_unicos)+1);
#ifdef VAR_RANGES
		logval("total_unicos", ((int)total_unicos));
#endif
		;
		_m = 3; goto P999; /* 0 */
	case 21: // STATE 13 - modelo_verificacion_spin.pml:71 - [en_critica = (en_critica-1)] (0:0:1 - 1)
		IfNotBlocked
		reached[1][13] = 1;
		(trpt+1)->bup.oval = ((int)now.en_critica);
		now.en_critica = (((int)now.en_critica)-1);
#ifdef VAR_RANGES
		logval("en_critica", ((int)now.en_critica));
#endif
		;
		_m = 3; goto P999; /* 0 */
	case 22: // STATE 14 - modelo_verificacion_spin.pml:25 - [mutex = 0] (0:0:1 - 1)
		IfNotBlocked
		reached[1][14] = 1;
		(trpt+1)->bup.oval = ((int)now.mutex);
		now.mutex = 0;
#ifdef VAR_RANGES
		logval("mutex", ((int)now.mutex));
#endif
		;
		_m = 3; goto P999; /* 0 */
	case 23: // STATE 16 - modelo_verificacion_spin.pml:75 - [procesado!t] (0:0:0 - 1)
		IfNotBlocked
		reached[1][16] = 1;
		if (q_full(now.procesado))
			continue;
#ifdef HAS_CODE
		if (readtrail && gui) {
			char simtmp[64];
			sprintf(simvals, "%d!", now.procesado);
		sprintf(simtmp, "%d", ((int)((P1 *)_this)->t)); strcat(simvals, simtmp);		}
#endif
		
		qsend(now.procesado, 0, ((int)((P1 *)_this)->t), 1);
		_m = 2; goto P999; /* 0 */
	case 24: // STATE 22 - modelo_verificacion_spin.pml:78 - [-end-] (0:0:0 - 3)
		IfNotBlocked
		reached[1][22] = 1;
		if (!delproc(1, II)) continue;
		_m = 3; goto P999; /* 0 */

		 /* PROC Productor */
	case 25: // STATE 1 - modelo_verificacion_spin.pml:32 - [((i<4))] (0:0:0 - 1)
		IfNotBlocked
		reached[0][1] = 1;
		if (!((((int)((P0 *)_this)->i)<4)))
			continue;
		_m = 3; goto P999; /* 0 */
	case 26: // STATE 2 - modelo_verificacion_spin.pml:33 - [tareas!i] (0:0:0 - 1)
		IfNotBlocked
		reached[0][2] = 1;
		if (q_full(now.tareas))
			continue;
#ifdef HAS_CODE
		if (readtrail && gui) {
			char simtmp[64];
			sprintf(simvals, "%d!", now.tareas);
		sprintf(simtmp, "%d", ((int)((P0 *)_this)->i)); strcat(simvals, simtmp);		}
#endif
		
		qsend(now.tareas, 0, ((int)((P0 *)_this)->i), 1);
		_m = 2; goto P999; /* 0 */
	case 27: // STATE 3 - modelo_verificacion_spin.pml:34 - [i = (i+1)] (0:0:1 - 1)
		IfNotBlocked
		reached[0][3] = 1;
		(trpt+1)->bup.oval = ((int)((P0 *)_this)->i);
		((P0 *)_this)->i = (((int)((P0 *)_this)->i)+1);
#ifdef VAR_RANGES
		logval("Productor:i", ((int)((P0 *)_this)->i));
#endif
		;
		_m = 3; goto P999; /* 0 */
	case 28: // STATE 9 - modelo_verificacion_spin.pml:39 - [i = 0] (0:15:1 - 3)
		IfNotBlocked
		reached[0][9] = 1;
		(trpt+1)->bup.oval = ((int)((P0 *)_this)->i);
		((P0 *)_this)->i = 0;
#ifdef VAR_RANGES
		logval("Productor:i", ((int)((P0 *)_this)->i));
#endif
		;
		/* merge: .(goto)(0, 16, 15) */
		reached[0][16] = 1;
		;
		_m = 3; goto P999; /* 1 */
	case 29: // STATE 10 - modelo_verificacion_spin.pml:41 - [((i<2))] (0:0:0 - 1)
		IfNotBlocked
		reached[0][10] = 1;
		if (!((((int)((P0 *)_this)->i)<2)))
			continue;
		_m = 3; goto P999; /* 0 */
	case 30: // STATE 11 - modelo_verificacion_spin.pml:42 - [tareas!255] (0:0:0 - 1)
		IfNotBlocked
		reached[0][11] = 1;
		if (q_full(now.tareas))
			continue;
#ifdef HAS_CODE
		if (readtrail && gui) {
			char simtmp[64];
			sprintf(simvals, "%d!", now.tareas);
		sprintf(simtmp, "%d", 255); strcat(simvals, simtmp);		}
#endif
		
		qsend(now.tareas, 0, 255, 1);
		_m = 2; goto P999; /* 0 */
	case 31: // STATE 12 - modelo_verificacion_spin.pml:43 - [i = (i+1)] (0:0:1 - 1)
		IfNotBlocked
		reached[0][12] = 1;
		(trpt+1)->bup.oval = ((int)((P0 *)_this)->i);
		((P0 *)_this)->i = (((int)((P0 *)_this)->i)+1);
#ifdef VAR_RANGES
		logval("Productor:i", ((int)((P0 *)_this)->i));
#endif
		;
		_m = 3; goto P999; /* 0 */
	case 32: // STATE 18 - modelo_verificacion_spin.pml:47 - [-end-] (0:0:0 - 3)
		IfNotBlocked
		reached[0][18] = 1;
		if (!delproc(1, II)) continue;
		_m = 3; goto P999; /* 0 */
	case  _T5:	/* np_ */
		if (!((!(trpt->o_pm&4) && !(trpt->tau&128))))
			continue;
		/* else fall through */
	case  _T2:	/* true */
		_m = 3; goto P999;
#undef rand
	}

