#define N 5
#define DATOS 5

chan buffer = [N] of { byte };

proctype Productor() {
    byte i = 0;

    do
    :: (i < DATOS) ->
        buffer!i;
        printf("Productor envia: %d\n", i);
        i++
    :: else ->
        break
    od
}

proctype Consumidor() {
    byte dato;

    do
    :: buffer?dato ->
        printf("Consumidor procesa: %d\n", dato)
    od
}

init {
    run Productor();
    run Consumidor();
}