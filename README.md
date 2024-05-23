# Saug

Grast mde Threads ab und lädt Bilder von abload (nicht-rip) herunter.

## So geht's

Passenden Release herunterladen. Irgendwo hinkopieren. Terminal öffnen.

Einzelnen Thread

```
./saug 196182
```

Mehrere Threads

```
./saug 123 345 678
```

## MacOS

Um böse binaries wie dieses hier auszuführen: Im Terminal wo das binary liegt (immer brav mit tab completen):

```
xattr -rd com.apple.quarantine saug
```

## MacOS und Linux

Natürlich muss das ganze ausführbar sein (auch hier, immer brav mit tab completen):

```
chmod +x saug
```

Für jede Threadid wird ein Ordner mit den Bildern angelegt. Von Thumbnails wird das Originalbild heruntergeladen. Viel Spass.
