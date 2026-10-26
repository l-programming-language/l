fn main() {
    assert_eq!(library_fixture::identity(7), 7);
    let value: library_fixture::Shared = library_fixture::make();
    let _: library_fixture::Shared = library_fixture::detail::preserve(value);
    let _: Option<&dyn library_fixture::Marker> = None;
}
