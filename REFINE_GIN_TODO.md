## Implementation Notes

1. Always check for compatibility with existing code
2. Don't create duplicate definitions
3. Follow existing naming conventions
4. Add proper documentation
5. Make sure all tests pass 

# Refine-Gin TODO - Plan rozszerzenia funkcjonalności

## 1. Wsparcie dla pól edytowalnych i tylko do odczytu ✅

- [x] Dodać pole `ReadOnly` do struktury `Field` w `pkg/resource/field.go`
- [x] Dodać pole `Hidden` do struktury `Field` do ukrywania pól w UI
- [x] Zmodyfikować `ParseFieldTag` aby obsługiwać tagi `readOnly` i `hidden`
- [x] Dodać tablicę `EditableFields` do `ResourceConfig` i `DefaultResource`
- [x] Zaimplementować metodę `GetEditableFields()` w interfejsie `Resource`
- [x] Zmodyfikować handlery operacji aktualizacji, aby respektowały status edytowalności pól

## 2. Zaawansowane wsparcie dla zagnieżdżonych struktur JSON

- [x] Rozszerzyć strukturę `JsonConfig` o:
  - [x] Pole `Nested` (bool) do odróżniania zagnieżdżonych struktur od płaskich JSON
  - [x] Pole `RenderAs` (string) określające sposób renderowania ("tabs", "form", "tree", "grid")
  - [x] Pole `TabsConfig` do konfiguracji prezentacji w formie zakładek
  - [x] Pole `GridConfig` do konfiguracji układu siatki
  - [x] Pole `ObjectLabels` (mapa) dla etykiet sekcji/obiektów

- [x] Dodać struktury konfiguracyjne dla zakładek:
  - [x] `JsonTabsConfig` ze spisem zakładek
  - [x] `JsonTab` z konfiguracją pojedynczej zakładki (tytuł, ikona, pola)

- [x] Dodać struktury konfiguracyjne dla układu siatki:
  - [x] `JsonGridConfig` do definiowania kolumn i układu
  - [x] `JsonFieldLayout` do pozycjonowania pól w siatce

- [x] Zaimplementować funkcję `ValidateNestedJson` do walidacji zagnieżdżonych struktur
- [x] Rozszerzyć generowanie metadanych dla pól JSON o nowe opcje
- [ ] Zmodyfikować handlery CRUD, aby obsługiwały walidację i przetwarzanie zagnieżdżonych struktur

## 3. Wsparcie dla specjalnych typów pól

- [x] Dodać obsługę pól typu `File`/`Image`:
  - [x] Stworzyć strukturę `FileConfig` w `pkg/resource/field.go`
  - [x] Dodać pola konfiguracyjne (typy MIME, maks. rozmiar, ścieżka zapisu, URL)
  - [x] Dodać wsparcie dla obrazów (miniatur, wymiarów, przetwarzania)

- [x] Dodać obsługę pól typu `RichText`:
  - [x] Stworzyć strukturę `RichTextConfig` w `pkg/resource/field.go`
  - [x] Dodać opcje konfiguracyjne (pasek narzędzi, limity, zachowanie)

- [x] Dodać rozszerzone wsparcie dla pól typu `Select`:
  - [x] Rozszerzyć strukturę `SelectConfig` o zaawansowane opcje (wiele wyborów, zależności)
  - [x] Dodać obsługę dynamicznych opcji (URL API)
  - [x] Dodać wsparcie dla autouzupełniania

- [x] Dodać wsparcie dla pól obliczanych (`ComputedField`):
  - [x] Stworzyć strukturę konfiguracyjną z definicją zależności
  - [x] Zaimplementować mechanizm obliczania wartości na podstawie innych pól

## 4. Dostosowanie do Ant Design

- [x] Dodać strukturę `AntDesignConfig` do `Field` z konfiguracją specyficzną dla Ant Design:
  - [x] Typ komponentu (np. "Input", "Select", "DatePicker")
  - [x] Właściwości przekazywane do komponentu (props)
  - [x] Reguły walidacji zgodne z formatem Ant Design Form

- [x] Dodać funkcję mapującą standardową walidację na reguły Ant Design
- [x] Zmodyfikować generator metadanych, aby zawierał informacje specyficzne dla Ant Design
- [x] Dodać automatyczne wykrywanie typu komponentu Ant Design na podstawie typu pola

## 5. System uprawnień na poziomie pól i zasobów

- [x] Dodać mapę `Permissions` do struktur `Field` i `ResourceConfig`:
  - [x] Klucze: operacje ("create", "read", "update", "delete")
  - [x] Wartości: tablice ról mających uprawnienia

- [x] Rozszerzyć metadane OPTIONS o informacje o uprawnieniach
- [x] Zaimplementować middleware do sprawdzania uprawnień na poziomie operacji CRUD
- [x] Zaimplementować filtrowanie pól w odpowiedziach na podstawie uprawnień

## 6. Walidacja i metadane walidacji ✅

- [x] Rozszerzyć strukturę `Validation` o dodatkowe pola:
  - [x] `Custom` (string) dla niestandardowych reguł
  - [x] `Conditional` dla walidacji warunkowej
  - [x] `AsyncValidator` (URL) dla walidacji asynchronicznej

- [x] Dodać szczegółowe metadane walidacyjne do odpowiedzi OPTIONS:
  - [x] Reguły walidacji dla każdego pola
  - [x] Komunikaty błędów
  - [x] Informacje dla walidacji po stronie klienta (np. licznik znaków)

## 7. Formularze i układ formularzy

- [x] Dodać strukturę `FormLayout` do konfiguracji układu formularzy:
  - [x] Grupowanie pól
  - [x] Definiowanie sekcji
  - [x] Wielokolumnowy układ
  - [ ] Ukryte sekcje rozwijane

- [ ] Rozszerzyć strukturę `FormConfig` o dodatkowe opcje:
  - [ ] `Width` (procent szerokości formularza)
  - [x] `Dependent` (zależność od innych pól)
  - [ ] `Condition` (warunek wyświetlania)

- [x] Zaimplementować endpoint dedykowany dla metadanych formularza, który zwraca:
  - [x] Pełną definicję układu
  - [x] Wstępnie wypełnione wartości
  - [x] Zależności między polami

## 8. Pre/post processing danych

- [ ] Dodać mechanizm hooks do zasobów:
  - [ ] `BeforeCreate`, `AfterCreate`
  - [ ] `BeforeUpdate`, `AfterUpdate`
  - [ ] `BeforeDelete`, `AfterDelete`

- [ ] Zaimplementować transformację danych podobną do Marshmallow:
  - [ ] `OnLoad` - przetwarzanie danych wejściowych
  - [ ] `OnDump` - przetwarzanie danych wyjściowych
  - [ ] Transformacje specyficzne dla pól

## 9. Integracje i narzędzia

- [ ] Dodać generator kodu dla typowych przypadków użycia:
  - [ ] Generowanie zasobów z modeli
  - [ ] Generowanie schematów JSON z struktur
  - [ ] Narzędzia CLI do tworzenia projektów

- [ ] Rozszerzyć dokumentację o przykłady integracji z refine.dev:
  - [ ] Przykłady konfiguracji dla różnych typów pól
  - [ ] Przykłady zagnieżdżonych formularzy
  - [ ] Przykłady zaawansowanej walidacji

## 10. Testy i jakość kodu

- [x] Dodać kompleksowe testy dla nowych funkcjonalności:
  - [x] Testy jednostkowe dla walidacji zagnieżdżonych struktur
  - [x] Testy dla metadanych specjalnych typów pól (File, RichText, Select, Computed)
  - [ ] Testy integracyjne dla widoków formularzy
  - [ ] Testy wydajnościowe dla złożonych struktur JSON

- [ ] Zoptymalizować wydajność:
  - [ ] Zaimplementować efektywniejsze mapowanie dla złożonych struktur
  - [ ] Dodać cache'owanie metadanych
  - [ ] Zoptymalizować walidację dla dużych zagnieżdżonych obiektów 