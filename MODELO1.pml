#define N 3   // número de consumidores (workers)

chan buffer = [5] of { int };

proctype Productor() {
    int i = 0;
    do
    :: i < 10 ->
        printf("Productor envia: %d\n", i);
        buffer!i;
        i++
    :: else ->
        break
    od;
}

// Consumidor parametrizado
proctype Consumidor(byte id) {
    int x;
    do
    :: buffer?x ->
        printf("Worker %d procesa: %d\n", id, x)
    od;
}

init {
    byte i = 0;

    // lanzar productor
    run Productor();

    // lanzar múltiples consumidores
    do
    :: i < N ->
        run Consumidor(i);
        i++
    :: else ->
        break
    od;
}