# Call of Duty Settings Patcher

Un outil Windows en terminal pour appliquer facilement des réglages Call of Duty sélectionnés à vos fichiers de configuration.

L’application détecte vos configurations, affiche les changements avant application et crée une sauvegarde automatique de chaque fichier modifié.

## Télécharger

Téléchargez la dernière archive Windows depuis la page [Releases](../../releases), extrayez-la dans le dossier de votre choix, puis lancez `cod-settings-patcher.exe`.

## Utilisation

1. **Fermez Call of Duty** avant de lancer l’outil.
2. Ouvrez un terminal dans le dossier contenant l’exécutable.
3. Lancez :

   ```powershell
   .\cod-settings-patcher.exe
   ```

4. Sélectionnez l’installation détectée avec `↑`/`↓` ou `j`/`k`.
5. Appuyez sur `Entrée` pour voir les changements proposés.
6. Vérifiez le jeu, le dossier, les fichiers et les valeurs qui vont changer.
7. Confirmez avec `y` pour appliquer les changements, ou annulez avec `n`/`Échap`.

L’outil ne crée pas de nouveaux réglages : il modifie uniquement les clés déjà présentes dans vos fichiers de configuration.

## Sauvegardes

Avant toute modification, une copie horodatée de chaque fichier est créée à côté du fichier d’origine. En cas de souci, restaurez simplement le fichier `.backup-...` correspondant.

## Avertissement SmartScreen

Windows peut afficher un avertissement **Microsoft Defender SmartScreen** pour un exécutable téléchargé depuis Internet qui n’est pas signé ou n’a pas encore acquis de réputation.

Ne contournez cet avertissement que si vous avez téléchargé l’archive depuis la page [Releases](../../releases) officielle de ce dépôt et, idéalement, vérifié son checksum SHA-256 via `checksums.txt`.

## Documentation avancée

Pour la liste complète des réglages, le détail de `RendererWorkerCount`, les sauvegardes, la compilation et les vérifications techniques, consultez [ADVANCED.md](ADVANCED.md).
