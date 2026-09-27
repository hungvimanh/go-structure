namespace CHECKIN.Entities
{
    public class Major : DataEntity
    {
        public long Id { get; set; }
        public string Code { get; set; }
        public string Name { get; set; }
        public long OrganizationId { get; set; }
        public long StatusId { get; set; }
        public Organization Organization { get; set; }
        public Status Status { get; set; }
        public List<MajorSubjectMapping> MajorSubjectMappings { get; set; }
        public Guid RowId { get; set; }
        public DateTime CreatedAt { get; set; }
        public DateTime UpdatedAt { get; set; }
        public DateTime? DeletedAt { get; set; }
        public bool Used { get; set; }
    }

    public class MajorFilter : FilterEntity
    {
        public IdFilter Id { get; set; }
        public StringFilter Code { get; set; }
        public StringFilter Name { get; set; }
        public IdFilter OrganizationId { get; set; }
        public IdFilter StatusId { get; set; }
        public DateFilter CreatedAt { get; set; }
        public DateFilter UpdatedAt { get; set; }
        public List<MajorFilter> OrFilter { get; set; }
        public MajorOrder OrderBy { get; set; }
        public MajorSelect Selects { get; set; } = MajorSelect.ALL;
        public MajorSearch SearchBy { get; set; }
    }

    [JsonConverter(typeof(StringEnumConverter))]
    public enum MajorOrder
    {
        Id = 0,
        Code = 1,
        Name = 2,
        Organization = 3,
        Status = 4,
        CreatedAt = 50,
        UpdatedAt = 51,
    }

    [Flags]
    public enum MajorSelect : long
    {
        ALL = E.ALL,
        Id = E._0,
        Code = E._1,
        Name = E._2,
        Organization = E._3,
        Status = E._4,
    }

    [Flags]
    public enum MajorSearch : long
    {
        ALL = E.ALL,
        Code = E._1,
        Name = E._2,
    }
}
