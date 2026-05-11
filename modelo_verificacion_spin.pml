#define N_TAREAS 4
#define N_WORKERS 2
#define SENAL_FIN 255

chan tareas = [2] of { byte };
chan procesado = [2] of { byte };

/*
  Variables compartidas para modelar la sección crítica.
  Representan el acceso al mapa de sumillas vistas y contadores globales
  protegidos con sync.Mutex en Go.
*/
bool mutex = false;
byte en_critica = 0;
byte total_unicos = 0;
byte total_consumidos = 0;

inline lock() {
    atomic {
        mutex == false -> mutex = true
    }
}

inline unlock() {
    mutex = false
}

proctype Productor() {
    byte i = 0;

    do
    :: i < N_TAREAS ->
        tareas!i;
        i++
    :: else ->
        break
    od;

    i = 0;
    do
    :: i < N_WORKERS ->
        tareas!SENAL_FIN;
        i++
    :: else ->
        break
    od
}

proctype Worker(byte id) {
    byte t;

    do
    :: tareas?t ->
        if
        :: t == SENAL_FIN ->
            procesado!SENAL_FIN;
            break

        :: else ->
            /*
              Sección crítica:
              simula deduplicación global y actualización de contadores.
            */
            lock();

            en_critica++;
            assert(en_critica == 1);

            total_unicos++;

            en_critica--;

            unlock();

            procesado!t
        fi
    od
}

proctype Consumidor() {
    byte r;
    byte workers_finalizados = 0;

    do
    :: procesado?r ->
        if
        :: r == SENAL_FIN ->
            workers_finalizados++;

            if
            :: workers_finalizados == N_WORKERS ->
                break
            :: else ->
                skip
            fi

        :: else ->
            total_consumidos++
        fi
    od;

    assert(total_consumidos == N_TAREAS)
}

init {
    atomic {
        run Productor();
        run Worker(0);
        run Worker(1);
        run Consumidor()
    }
}