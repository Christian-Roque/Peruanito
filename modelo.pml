chan tareas = [5] of {int};
chan procesado = [5] of {int};

active proctype Productor() {
    int i = 0;
    do
    :: i < 10 ->
        printf("Productor envia texto RAW %d\n", i);
        tareas!i;
        i++
    :: else -> break
    od
}

active proctype Worker() {
    int t;
    do
    :: tareas?t ->
        printf("Worker recibe texto %d\n", t);

        /* Simulación del pipeline NLP */
        printf(" - limpiando texto\n");
        printf(" - tokenizando\n");
        printf(" - eliminando stopwords\n");

        procesado!t
    od
}

active proctype Consumidor() {
    int r;
    do
    :: procesado?r ->
        printf("Consumidor recibe texto LIMPIO %d\n", r)
    od
}

init {
    run Productor();
    run Worker();
    run Worker();
    run Worker(); // más concurrencia
    run Consumidor();
}